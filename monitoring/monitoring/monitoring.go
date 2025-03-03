package monitoring

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/grafana-tools/sdk"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring/internal/grafana"
)

// Container describes a Docker container to be observed.
//
// These correspond to dashboards in Grafana.
type Container struct {
	// Name of the Docker container, e.g. "syntect-server".
	Name string

	// Title of the Docker container, e.g. "Syntect Server".
	Title string

	// Description of the Docker container. It should describe what the container
	// is responsible for, so that the impact of issues in it is clear.
	Description string

	// Variables define the variables that can be to applied to the dashboard for this
	// container, such as instances or shards.
	Variables []ContainerVariable

	// RawVariables is an alternative to Variables that exposes the underlying Grafana API
	// to define variables that can be applied to the dashboard for this container.
	//
	// It is recommended to use or expand the standardized Variables field instead.
	RawVariables []sdk.TemplateVar

	// Groups of observable information about the container.
	Groups []Group

	// NoSourcegraphDebugServer indicates if this container does not export the standard
	// Sourcegraph debug server (package `internal/debugserver`).
	//
	// This is used to configure monitoring features that depend on information exported
	// by the standard Sourcegraph debug server.
	NoSourcegraphDebugServer bool
}

func (c *Container) validate() error {
	if !isValidGrafanaUID(c.Name) {
		return errors.Errorf("Name must be lowercase alphanumeric + dashes; found \"%s\"", c.Name)
	}
	if c.Title != strings.Title(c.Title) {
		return errors.Errorf("Title must be in Title Case; found \"%s\" want \"%s\"", c.Title, strings.Title(c.Title))
	}
	if c.Description != withPeriod(c.Description) || c.Description != upperFirst(c.Description) {
		return errors.Errorf("Description must be sentence starting with an uppercase letter and ending with period; found \"%s\"", c.Description)
	}
	for i, v := range c.Variables {
		if err := v.validate(); err != nil {
			return errors.Errorf("Variable %d %q: %v", i, c.Name, err)
		}
	}
	for i, g := range c.Groups {
		if err := g.validate(); err != nil {
			return errors.Errorf("Group %d %q: %v", i, g.Title, err)
		}
	}
	return nil
}

// renderDashboard generates the Grafana renderDashboard for this container.
func (c *Container) renderDashboard() *sdk.Board {
	board := sdk.NewBoard(c.Title)
	board.Version = uint(rand.Uint32())
	board.UID = c.Name
	board.ID = 0
	board.Timezone = "utc"
	board.Timepicker.RefreshIntervals = []string{"5s", "10s", "30s", "1m", "5m", "15m", "30m", "1h", "2h", "1d"}
	board.Time.From = "now-6h"
	board.Time.To = "now"
	board.SharedCrosshair = true
	board.Editable = false
	board.AddTags("builtin")
	alertLevelVariable := ContainerVariable{
		Label:   "Alert level",
		Name:    "alert_level",
		Options: []string{"critical", "warning"},
	}
	board.Templating.List = []sdk.TemplateVar{alertLevelVariable.toGrafanaTemplateVar()}
	for _, variable := range c.Variables {
		board.Templating.List = append(board.Templating.List, variable.toGrafanaTemplateVar())
	}
	board.Templating.List = append(board.Templating.List, c.RawVariables...)
	board.Annotations.List = []sdk.Annotation{{
		Name:       "Alert events",
		Datasource: StringPtr("Prometheus"),
		// Show alerts matching the selected alert_level (see template variable above)
		Expr:        fmt.Sprintf(`ALERTS{service_name=%q,level=~"$alert_level",alertstate="firing"}`, c.Name),
		Step:        "60s",
		TitleFormat: "{{ description }} ({{ name }})",
		TagKeys:     "level,owner",
		IconColor:   "rgba(255, 96, 96, 1)",
		Enable:      false, // disable by default for now
		Type:        "tags",
	}}
	// Annotation layers that require a service to export information required by the
	// Sourcegraph debug server - see the `NoSourcegraphDebugServer` docstring.
	if !c.NoSourcegraphDebugServer {
		board.Annotations.List = append(board.Annotations.List, sdk.Annotation{
			Name:       "Version changes",
			Datasource: StringPtr("Prometheus"),
			// Per version, instance generate an annotation whenever labels change
			// inspired by https://github.com/grafana/grafana/issues/11948#issuecomment-403841249
			// We use `job=~.*SERVICE` because of frontend being called sourcegraph-frontend in certain environments
			Expr:        fmt.Sprintf(`group by(version, instance) (src_service_metadata{job=~".*%[1]s"} unless (src_service_metadata{job=~".*%[1]s"} offset 1m))`, c.Name),
			Step:        "60s",
			TitleFormat: "v{{ version }}",
			TagKeys:     "instance",
			IconColor:   "rgb(255, 255, 255)",
			Enable:      false, // disable by default for now
			Type:        "tags",
		})
	}

	description := sdk.NewText("")
	description.Title = "" // Removes vertical space the title would otherwise take up
	setPanelSize(description, 24, 3)
	description.TextPanel.Mode = "html"
	description.TextPanel.Content = fmt.Sprintf(`
	<div style="text-align: left;">
	  <img src="https://sourcegraphstatic.com/sourcegraph-logo-light.png" style="height:30px; margin:0.5rem"></img>
	  <div style="margin-left: 1rem; margin-top: 0.5rem; font-size: 20px;"><b>%s:</b> %s <a style="font-size: 15px" target="_blank" href="https://docs.sourcegraph.com/dev/background-information/architecture">(⧉ architecture diagram)</a></span>
	</div>
	`, c.Name, c.Description)
	board.Panels = append(board.Panels, description)

	alertsDefined := grafana.NewContainerAlertsDefinedTable(sdk.Target{
		Expr: fmt.Sprintf(`label_replace(
			sum(max by (level,service_name,name,description,grafana_panel_id)(alert_count{service_name="%s",name!="",level=~"$alert_level"})) by (level,description,service_name,grafana_panel_id),
			"description", "$1", "description", ".*: (.*)"
		)`, c.Name),
		Format:  "table",
		Instant: true,
	})
	setPanelSize(alertsDefined, 9, 5)
	setPanelPos(alertsDefined, 0, 3)
	board.Panels = append(board.Panels, alertsDefined)

	alertsFiring := sdk.NewGraph("Alerts firing")
	setPanelSize(alertsFiring, 15, 5)
	setPanelPos(alertsFiring, 9, 3)
	alertsFiring.GraphPanel.Legend.Show = true
	alertsFiring.GraphPanel.Fill = 1
	alertsFiring.GraphPanel.Bars = true
	alertsFiring.GraphPanel.NullPointMode = "null"
	alertsFiring.GraphPanel.Pointradius = 2
	alertsFiring.GraphPanel.AliasColors = map[string]string{}
	alertsFiring.GraphPanel.Xaxis = sdk.Axis{
		Show: true,
	}
	alertsFiring.GraphPanel.Yaxes = []sdk.Axis{
		{
			Decimals: 0,
			Format:   "short",
			LogBase:  1,
			Max:      sdk.NewFloatString(1),
			Min:      sdk.NewFloatString(0),
			Show:     false,
		},
		{
			Format:  "short",
			LogBase: 1,
			Show:    true,
		},
	}
	alertsFiring.AddTarget(&sdk.Target{
		Expr:         fmt.Sprintf(`sum by (service_name,level,name,grafana_panel_id)(max by (level,service_name,name,description,grafana_panel_id)(alert_count{service_name="%s",name!="",level=~"$alert_level"}) >= 1)`, c.Name),
		LegendFormat: "{{level}}: {{name}}",
	})
	alertsFiring.GraphPanel.FieldConfig = &sdk.FieldConfig{}
	alertsFiring.GraphPanel.FieldConfig.Defaults.Links = []sdk.Link{{
		Title: "Graph panel",
		URL:   StringPtr("/-/debug/grafana/d/${__field.labels.service_name}/${__field.labels.service_name}?viewPanel=${__field.labels.grafana_panel_id}"),
	}}
	board.Panels = append(board.Panels, alertsFiring)

	baseY := 8
	offsetY := baseY
	for groupIndex, group := range c.Groups {
		// Non-general groups are shown as collapsible panels.
		var rowPanel *sdk.Panel
		if group.Title != "General" {
			rowPanel = &sdk.Panel{RowPanel: &sdk.RowPanel{}}
			rowPanel.OfType = sdk.RowType
			rowPanel.Type = "row"
			rowPanel.Title = group.Title
			offsetY++
			setPanelPos(rowPanel, 0, offsetY)
			rowPanel.Collapsed = group.Hidden
			rowPanel.Panels = []sdk.Panel{} // cannot be null
			board.Panels = append(board.Panels, rowPanel)
		}

		// Generate a panel for displaying each observable in each row.
		for rowIndex, row := range group.Rows {
			panelWidth := 24 / len(row)
			offsetY++
			for i, o := range row {
				panelTitle := strings.ToTitle(string([]rune(o.Description)[0])) + string([]rune(o.Description)[1:])

				var panel *sdk.Panel
				switch o.Panel.panelType {
				case PanelTypeGraph:
					panel = sdk.NewGraph(panelTitle)
				case PanelTypeHeatmap:
					panel = sdk.NewHeatmap(panelTitle)
				}

				panel.ID = observablePanelID(groupIndex, rowIndex, i)

				// Set positioning
				setPanelSize(panel, panelWidth, 5)
				setPanelPos(panel, i*panelWidth, offsetY)

				// Add reference links
				panel.Links = []sdk.Link{{
					Title:       "Panel reference",
					URL:         StringPtr(fmt.Sprintf("%s#%s", canonicalDashboardsDocsURL, observableDocAnchor(c, o))),
					TargetBlank: boolPtr(true),
				}}
				if !o.NoAlert {
					panel.Links = append(panel.Links, sdk.Link{
						Title:       "Alerts reference",
						URL:         StringPtr(fmt.Sprintf("%s#%s", canonicalAlertSolutionsURL, observableDocAnchor(c, o))),
						TargetBlank: boolPtr(true),
					})
				}

				// Build the graph panel
				o.Panel.build(o, panel)

				// Attach panel to board
				if rowPanel != nil && group.Hidden {
					rowPanel.RowPanel.Panels = append(rowPanel.RowPanel.Panels, *panel)
				} else {
					board.Panels = append(board.Panels, panel)
				}
			}
		}
	}
	return board
}

// alertDescription generates an alert description for the specified coontainer's alert.
func (c *Container) alertDescription(o Observable, alert *ObservableAlertDefinition) (string, error) {
	if alert.isEmpty() {
		return "", errors.New("cannot generate description for empty alert")
	}
	var description string

	// description based on thresholds. no special description for 'alert.strictCompare',
	// because the description is pretty ambiguous to fit different alerts.
	units := o.Panel.unitType.short()
	if alert.greaterThan {
		// e.g. "zoekt-indexserver: 20+ indexed search request errors every 5m by code"
		description = fmt.Sprintf("%s: %v%s+ %s", c.Name, alert.threshold, units, o.Description)
	} else if alert.lessThan {
		// e.g. "zoekt-indexserver: less than 20 indexed search requests every 5m by code"
		description = fmt.Sprintf("%s: less than %v%s %s", c.Name, alert.threshold, units, o.Description)
	} else {
		return "", errors.Errorf("unable to generate description for observable %+v", o)
	}

	// add information about "for"
	if alert.duration > 0 {
		return fmt.Sprintf("%s for %s", description, alert.duration), nil
	}
	return description, nil
}

// renderRules generates the Prometheus rules file which defines our
// high-level alerting metrics for the container. For more information about
// how these work, see:
//
// https://docs.sourcegraph.com/admin/observability/metrics#high-level-alerting-metrics
//
func (c *Container) renderRules() (*promRulesFile, error) {
	group := promGroup{Name: c.Name}
	for groupIndex, g := range c.Groups {
		for rowIndex, r := range g.Rows {
			for observableIndex, o := range r {
				for level, a := range map[string]*ObservableAlertDefinition{
					"warning":  o.Warning,
					"critical": o.Critical,
				} {
					if a.isEmpty() {
						continue
					}

					// The alertQuery must contribute a query that returns true when it should be firing.
					alertQuery := fmt.Sprintf("%s((%s) %s %v)",
						a.aggregator, o.Query, a.comparator, a.threshold)

					// If the data must exist, we alert if the query returns no value as well
					if o.DataMustExist {
						alertQuery = fmt.Sprintf("(%s) OR (absent(%s) == 1)", alertQuery, o.Query)
					}

					// Build the rule with appropriate labels. Labels are leveraged in various integrations, such as with prom-wrapper.
					description, err := c.alertDescription(o, a)
					if err != nil {
						return nil, errors.Errorf("%s.%s.%s: unable to generate labels: %+v",
							c.Name, o.Name, level, err)
					}
					group.appendRow(alertQuery, map[string]string{
						"name":         o.Name,
						"level":        level,
						"service_name": c.Name,
						"description":  description,
						"owner":        string(o.Owner),

						// in the corresponding dashboard, this label should indicate
						// the panel associated with this rule
						"grafana_panel_id": strconv.Itoa(int(observablePanelID(groupIndex, rowIndex, observableIndex))),
					}, a.duration)
				}
			}
		}
	}
	if err := group.validate(); err != nil {
		return nil, err
	}
	return &promRulesFile{
		Groups: []promGroup{group},
	}, nil
}

// ContainerVariable describes a template variable that can be applied container dashboard
// for filtering purposes.
type ContainerVariable struct {
	// Name is the name of the variable to substitute the value for, e.g. "alert_level"
	// to replace "$alert_level" in queries
	Name string
	// Label is a human-readable name for the variable, e.g. "Alert level"
	Label string

	// OptionsQuery is the query to generate the possible values for this variable. Cannot
	// be used in conjunction with Options
	OptionsQuery string
	// Options are the pre-defined possible values for this variable. Cannot be used in
	// conjunction with Query
	Options []string

	// WildcardAllValue indicates to Grafana that is should NOT use Query or Options to
	// generate a concatonated 'All' value for the variable, and use a '.*' wildcard
	// instead. Setting this to true primarily useful if you use Query and expect it to be
	// a large enough result set to cause issues when viewing the dashboard.
	//
	// We allow Grafana to generate a value by default because simply using '.*' wildcard
	// can pull in unintended metrics if adequate filtering is not performed on the query,
	// for example if multiple services export the same metric. If set to true, make sure
	// the queries that use this variable perform adequate filtering to avoid pulling in
	// unintended metrics.
	WildcardAllValue bool

	// Multi indicates whether or not to allow multi-selection for this variable filter
	Multi bool
}

func (c *ContainerVariable) validate() error {
	if c.Name == "" {
		return errors.New("ContainerVariable.Name is required")
	}
	if c.Label == "" {
		return errors.New("ContainerVariable.Label is required")
	}
	if c.OptionsQuery == "" && len(c.Options) == 0 {
		return errors.New("ContainerVariable.Query and ContainerVariable.Options cannot both be set")
	}
	return nil
}

// toGrafanaTemplateVar generates the Grafana template variable configuration for this
// container variable.
func (c *ContainerVariable) toGrafanaTemplateVar() sdk.TemplateVar {
	variable := sdk.TemplateVar{
		Name:  c.Name,
		Label: c.Label,
		Multi: c.Multi,

		Datasource: StringPtr("Prometheus"),
		IncludeAll: true,

		// Apply the AllValue to a template variable by default
		Current: sdk.Current{Text: &sdk.StringSliceString{Value: []string{"all"}, Valid: true}, Value: "$__all"},
	}

	if c.WildcardAllValue {
		variable.AllValue = ".*"
	} else {
		// Rely on Grafana to create a union of only the values
		// generated by the specified query.
		//
		// See https://grafana.com/docs/grafana/latest/variables/formatting-multi-value-variables/#multi-value-variables-with-a-prometheus-or-influxdb-data-source
		// for more information.
		variable.AllValue = ""
	}

	switch {
	case c.OptionsQuery != "":
		variable.Type = "query"
		variable.Query = c.OptionsQuery
		variable.Refresh = sdk.BoolInt{
			Flag:  true,
			Value: Int64Ptr(2), // Refresh on time range change
		}
		variable.Sort = 3

	case len(c.Options) > 0:
		variable.Type = "custom"
		variable.Query = strings.Join(c.Options, ",")
		// Add the AllValue as a default
		variable.Options = []sdk.Option{{Text: "all", Value: "$__all", Selected: true}}
		for _, option := range c.Options {
			variable.Options = append(variable.Options, sdk.Option{Text: option, Value: option})
		}
	}

	return variable
}

// Group describes a group of observable information about a container.
//
// These correspond to collapsible sections in a Grafana dashboard.
type Group struct {
	// Title of the group, briefly summarizing what this group is about, or
	// "General" if the group is just about the container in general.
	Title string

	// Hidden indicates whether or not the group should be hidden by default.
	//
	// This should only be used when the dashboard is already full of information
	// and the information presented in this group is unlikely to be the cause of
	// issues and should generally only be inspected in the event that an alert
	// for that information is firing.
	Hidden bool

	// Rows of observable metrics.
	Rows []Row
}

func (g Group) validate() error {
	if g.Title != upperFirst(g.Title) || g.Title == withPeriod(g.Title) {
		return errors.Errorf("Title must start with an uppercase letter and not end with a period; found \"%s\"", g.Title)
	}
	for i, r := range g.Rows {
		if err := r.validate(); err != nil {
			return errors.Errorf("Row %d: %v", i, err)
		}
	}
	return nil
}

// Row of observable metrics.
//
// These correspond to a row of Grafana graphs.
type Row []Observable

func (r Row) validate() error {
	if len(r) < 1 || len(r) > 4 {
		return errors.Errorf("row must have 1 to 4 observables only, found %v", len(r))
	}
	for i, o := range r {
		if err := o.validate(); err != nil {
			return errors.Errorf("Observable %d %q: %v", i, o.Name, err)
		}
	}
	return nil
}

// ObservableOwner denotes a team that owns an Observable. The current teams are described in
// the handbook: https://handbook.sourcegraph.com/engineering/eng_org#current-organization
type ObservableOwner string

const (
	ObservableOwnerSearch          ObservableOwner = "search"
	ObservableOwnerSearchCore      ObservableOwner = "search-core"
	ObservableOwnerBatches         ObservableOwner = "batches"
	ObservableOwnerCodeIntel       ObservableOwner = "code-intel"
	ObservableOwnerSecurity        ObservableOwner = "security"
	ObservableOwnerWeb             ObservableOwner = "web"
	ObservableOwnerCoreApplication ObservableOwner = "core application"
	ObservableOwnerCodeInsights    ObservableOwner = "code-insights"
	ObservableOwnerDevOps          ObservableOwner = "devops"
)

// toMarkdown returns a Markdown string that also links to the owner's team page
func (o ObservableOwner) toMarkdown() string {
	var slug string
	// special cases for differences in how a team is named in ObservableOwner and how
	// they are named in the handbook.
	// see https://handbook.sourcegraph.com/engineering/eng_org#current-organization
	switch o {
	case ObservableOwnerCodeIntel:
		slug = "code-intelligence"
	case ObservableOwnerCodeInsights:
		slug = "developer-insights/code-insights"
	case ObservableOwnerDevOps:
		slug = "cloud/devops"
	case ObservableOwnerSearchCore:
		slug = "search/core"
	default:
		slug = strings.ReplaceAll(string(o), " ", "-")
	}

	return fmt.Sprintf("[Sourcegraph %s team](https://handbook.sourcegraph.com/engineering/%s)",
		upperFirst(string(o)), slug)
}

// Observable describes a metric about a container that can be observed. For example, memory usage.
//
// These correspond to Grafana graphs.
type Observable struct {
	// Name is a short and human-readable lower_snake_case name describing what is being observed.
	//
	// It must be unique relative to the service name.
	//
	// Good examples:
	//
	//  github_rate_limit_remaining
	// 	search_error_rate
	//
	// Bad examples:
	//
	//  repo_updater_github_rate_limit
	// 	search_error_rate_over_5m
	//
	Name string

	// Description is a human-readable description of exactly what is being observed.
	// If a query groups by a label (such as with a `sum by(...)`), ensure that this is
	// reflected in the description by noting that this observable is grouped "by ...".
	//
	// Good examples:
	//
	// 	"remaining GitHub API rate limit quota"
	// 	"number of search errors every 5m"
	//  "90th percentile search request duration over 5m"
	//  "internal API error responses every 5m by route"
	//
	// Bad examples:
	//
	// 	"GitHub rate limit"
	// 	"search errors[5m]"
	// 	"P90 search latency"
	//
	Description string

	// Owner indicates the team that owns this Observable (including its alerts and maintainence).
	Owner ObservableOwner

	// Query is the actual Prometheus query that should be observed.
	Query string

	// DataMustExist indicates if the query must return data.
	//
	// For example, repo_updater_memory_usage should always have data present and an alert should
	// fire if for some reason that query is not returning any data, so this would be set to true.
	// In contrast, search_error_rate would depend on users actually performing searches and we
	// would not want an alert to fire if no data was present, so this will not need to be set.
	DataMustExist bool

	// Warning and Critical alert definitions.
	// Consider adding at least a Warning or Critical alert to each Observable to make it
	// easy to identify when the target of this metric is misbehaving. If no alerts are
	// provided, NoAlert must be set and Interpretation must be provided.
	Warning, Critical *ObservableAlertDefinition

	// NoAlerts must be set by Observables that do not have any alerts.
	// This ensures the omission of alerts is intentional. If set to true, an Interpretation
	// must be provided in place of PossibleSolutions.
	NoAlert bool

	// PossibleSolutions is Markdown describing possible solutions in the event that the
	// alert is firing. This field not required if no alerts are attached to this Observable.
	// If there is no clear potential resolution or there is no alert configured, "none"
	// must be explicitly stated.
	//
	// Use the Interpretation field for additional guidance on understanding this Observable
	// that isn't directly related to solving it.
	//
	// Contacting support should not be mentioned as part of a possible solution, as it is
	// communicated elsewhere.
	//
	// To make writing the Markdown more friendly in Go, string literals like this:
	//
	// 	Observable{
	// 		PossibleSolutions: `
	// 			- Foobar 'some code'
	// 		`
	// 	}
	//
	// Becomes:
	//
	// 	- Foobar `some code`
	//
	// In other words:
	//
	// 1. The preceding newline is removed.
	// 2. The indentation in the string literal is removed (based on the last line).
	// 3. Single quotes become backticks.
	// 4. The last line (which is all indention) is removed.
	// 5. Non-list items are converted to a list.
	//
	PossibleSolutions string

	// Interpretation is Markdown that can serve as a reference for interpreting this
	// observable. For example, Interpretation could provide guidance on what sort of
	// patterns to look for in the observable's graph and document why this observable is
	// useful.
	//
	// If no alerts are configured for an observable, this field is required. If the
	// Description is sufficient to capture what this Observable describes, "none" must be
	// explicitly stated.
	//
	// To make writing the Markdown more friendly in Go, string literal processing as
	// PossibleSolutions is provided, though the output is not converted to a list.
	Interpretation string

	// Panel provides options for how to render the metric in the Grafana panel.
	// A recommended set of options and customizations are available from the `Panel()`
	// constructor.
	//
	// Additional customizations can be made via `ObservablePanel.With()` for cases where
	// the provided `ObservablePanel` is insufficient - see `ObservablePanelOption` for
	// more details.
	Panel ObservablePanel
}

func (o Observable) validate() error {
	if strings.Contains(o.Name, " ") || strings.ToLower(o.Name) != o.Name {
		return errors.Errorf("Name must be in lower_snake_case; found \"%s\"", o.Name)
	}
	if len(o.Description) == 0 {
		return errors.New("Description must be set")
	}
	if first, second := string([]rune(o.Description)[0]), string([]rune(o.Description)[1]); first != strings.ToLower(first) && second == strings.ToLower(second) {
		return errors.Errorf("Description must be lowercase except for acronyms; found \"%s\"", o.Description)
	}
	if o.Owner == "" && !o.NoAlert {
		return errors.New("Owner must be defined for observables with alerts")
	}
	if !o.Panel.panelType.validate() {
		return errors.New(`Panel.panelType must be "graph" or "heatmap"`)
	}

	allAlertsEmpty := o.alertsCount() == 0
	if allAlertsEmpty || o.NoAlert {
		// Ensure lack of alerts is intentional
		if allAlertsEmpty && !o.NoAlert {
			return errors.Errorf("Warning or Critical must be set or explicitly disable alerts with NoAlert")
		} else if !allAlertsEmpty && o.NoAlert {
			return errors.Errorf("An alert is set, but NoAlert is also true")
		}
		// PossibleSolutions if there are no alerts is redundant and likely an error
		if o.PossibleSolutions != "" {
			return errors.Errorf(`PossibleSolutions is not required if no alerts are configured - did you mean to provide an Interpretation instead?`)
		}
		// Interpretation must be provided and valid
		if o.Interpretation == "" {
			return errors.Errorf("Interpretation must be provided if no alerts are set")
		} else if o.Interpretation != "none" {
			if _, err := toMarkdown(o.Interpretation, false); err != nil {
				return errors.Errorf("Interpretation cannot be converted to Markdown: %w", err)
			}
		}
	} else {
		// Ensure alerts are valid
		for alertLevel, alert := range map[string]*ObservableAlertDefinition{
			"Warning":  o.Warning,
			"Critical": o.Critical,
		} {
			if err := alert.validate(); err != nil {
				return errors.Errorf("%s Alert: %w", alertLevel, err)
			}
		}
		// PossibleSolutions must be provided and valid
		if o.PossibleSolutions == "" {
			return errors.Errorf(`PossibleSolutions must list solutions or an explicit "none"`)
		} else if o.PossibleSolutions != "none" {
			if solutions, err := toMarkdown(o.PossibleSolutions, true); err != nil {
				return errors.Errorf("PossibleSolutions cannot be converted to Markdown: %w", err)
			} else if l := strings.ToLower(solutions); strings.Contains(l, "contact support") || strings.Contains(l, "contact us") {
				return errors.Errorf("PossibleSolutions should not include mentions of contacting support")
			}
		}
	}

	return nil
}

func (o Observable) alertsCount() (count int) {
	if !o.Warning.isEmpty() {
		count++
	}
	if !o.Critical.isEmpty() {
		count++
	}
	return
}

// Alert provides a builder for defining alerting on an Observable.
func Alert() *ObservableAlertDefinition {
	return &ObservableAlertDefinition{}
}

// ObservableAlertDefinition defines when an alert would be considered firing.
type ObservableAlertDefinition struct {
	greaterThan bool
	lessThan    bool
	duration    time.Duration
	// Wrap the query in `max()` or `min()` so that if there are multiple series (e.g. per-container)
	// they get "flattened" into a single metric. The `aggregator` variable sets the required operator.
	//
	// We only support per-service alerts, not per-container/replica, and not doing so can cause issues.
	// See https://github.com/sourcegraph/sourcegraph/issues/11571#issuecomment-654571953,
	// https://github.com/sourcegraph/sourcegraph/issues/17599, and related pull requests.
	aggregator string
	// Comparator sets how a metric should be compared against a threshold
	comparator string
	// Threshold sets the value to be compared against
	threshold float64
}

// GreaterOrEqual indicates the alert should fire when greater or equal the given value.
func (a *ObservableAlertDefinition) GreaterOrEqual(f float64, aggregator *string) *ObservableAlertDefinition {
	a.greaterThan = true
	if aggregator != nil {
		a.aggregator = *aggregator
	} else {
		a.aggregator = "max"
	}
	a.comparator = ">="
	a.threshold = f
	return a
}

// LessOrEqual indicates the alert should fire when less than or equal to the given value.
func (a *ObservableAlertDefinition) LessOrEqual(f float64, aggregator *string) *ObservableAlertDefinition {
	a.lessThan = true
	if aggregator != nil {
		a.aggregator = *aggregator
	} else {
		a.aggregator = "min"
	}
	a.comparator = "<="
	a.threshold = f
	return a
}

// Greater indicates the alert should fire when strictly greater to this value.
func (a *ObservableAlertDefinition) Greater(f float64, aggregator *string) *ObservableAlertDefinition {
	a.greaterThan = true
	if aggregator != nil {
		a.aggregator = *aggregator
	} else {
		a.aggregator = "max"
	}
	a.comparator = ">"
	a.threshold = f
	return a
}

// Less indicates the alert should fire when strictly less than this value.
func (a *ObservableAlertDefinition) Less(f float64, aggregator *string) *ObservableAlertDefinition {
	a.lessThan = true
	if aggregator != nil {
		a.aggregator = *aggregator
	} else {
		a.aggregator = "min"
	}
	a.comparator = "<"
	a.threshold = f
	return a
}

// For indicates how long the given thresholds must be exceeded for this alert to be
// considered firing. Defaults to 0s (immediately alerts when threshold is exceeded).
func (a *ObservableAlertDefinition) For(d time.Duration) *ObservableAlertDefinition {
	a.duration = d
	return a
}

func (a *ObservableAlertDefinition) isEmpty() bool {
	return a == nil || (*a == ObservableAlertDefinition{}) || (!a.greaterThan && !a.lessThan)
}

func (a *ObservableAlertDefinition) validate() error {
	if a.isEmpty() {
		return nil
	}
	if a.greaterThan && a.lessThan {
		return errors.New("only one bound (greater or less) can be set")
	}
	return nil
}
