// Code generated by go-mockgen 1.1.2; DO NOT EDIT.

package runner

import (
	"context"
	"sync"

	definition "github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	storetypes "github.com/sourcegraph/sourcegraph/internal/database/migration/storetypes"
)

// MockStore is a mock implementation of the Store interface (from the
// package
// github.com/sourcegraph/sourcegraph/internal/database/migration/runner)
// used for unit testing.
type MockStore struct {
	// DoneFunc is an instance of a mock function object controlling the
	// behavior of the method Done.
	DoneFunc *StoreDoneFunc
	// DownFunc is an instance of a mock function object controlling the
	// behavior of the method Down.
	DownFunc *StoreDownFunc
	// IndexStatusFunc is an instance of a mock function object controlling
	// the behavior of the method IndexStatus.
	IndexStatusFunc *StoreIndexStatusFunc
	// TransactFunc is an instance of a mock function object controlling the
	// behavior of the method Transact.
	TransactFunc *StoreTransactFunc
	// TryLockFunc is an instance of a mock function object controlling the
	// behavior of the method TryLock.
	TryLockFunc *StoreTryLockFunc
	// UpFunc is an instance of a mock function object controlling the
	// behavior of the method Up.
	UpFunc *StoreUpFunc
	// VersionsFunc is an instance of a mock function object controlling the
	// behavior of the method Versions.
	VersionsFunc *StoreVersionsFunc
	// WithMigrationLogFunc is an instance of a mock function object
	// controlling the behavior of the method WithMigrationLog.
	WithMigrationLogFunc *StoreWithMigrationLogFunc
}

// NewMockStore creates a new mock of the Store interface. All methods
// return zero values for all results, unless overwritten.
func NewMockStore() *MockStore {
	return &MockStore{
		DoneFunc: &StoreDoneFunc{
			defaultHook: func(error) error {
				return nil
			},
		},
		DownFunc: &StoreDownFunc{
			defaultHook: func(context.Context, definition.Definition) error {
				return nil
			},
		},
		IndexStatusFunc: &StoreIndexStatusFunc{
			defaultHook: func(context.Context, string, string) (storetypes.IndexStatus, bool, error) {
				return storetypes.IndexStatus{}, false, nil
			},
		},
		TransactFunc: &StoreTransactFunc{
			defaultHook: func(context.Context) (Store, error) {
				return nil, nil
			},
		},
		TryLockFunc: &StoreTryLockFunc{
			defaultHook: func(context.Context) (bool, func(err error) error, error) {
				return false, nil, nil
			},
		},
		UpFunc: &StoreUpFunc{
			defaultHook: func(context.Context, definition.Definition) error {
				return nil
			},
		},
		VersionsFunc: &StoreVersionsFunc{
			defaultHook: func(context.Context) ([]int, []int, []int, error) {
				return nil, nil, nil, nil
			},
		},
		WithMigrationLogFunc: &StoreWithMigrationLogFunc{
			defaultHook: func(context.Context, definition.Definition, bool, func() error) error {
				return nil
			},
		},
	}
}

// NewStrictMockStore creates a new mock of the Store interface. All methods
// panic on invocation, unless overwritten.
func NewStrictMockStore() *MockStore {
	return &MockStore{
		DoneFunc: &StoreDoneFunc{
			defaultHook: func(error) error {
				panic("unexpected invocation of MockStore.Done")
			},
		},
		DownFunc: &StoreDownFunc{
			defaultHook: func(context.Context, definition.Definition) error {
				panic("unexpected invocation of MockStore.Down")
			},
		},
		IndexStatusFunc: &StoreIndexStatusFunc{
			defaultHook: func(context.Context, string, string) (storetypes.IndexStatus, bool, error) {
				panic("unexpected invocation of MockStore.IndexStatus")
			},
		},
		TransactFunc: &StoreTransactFunc{
			defaultHook: func(context.Context) (Store, error) {
				panic("unexpected invocation of MockStore.Transact")
			},
		},
		TryLockFunc: &StoreTryLockFunc{
			defaultHook: func(context.Context) (bool, func(err error) error, error) {
				panic("unexpected invocation of MockStore.TryLock")
			},
		},
		UpFunc: &StoreUpFunc{
			defaultHook: func(context.Context, definition.Definition) error {
				panic("unexpected invocation of MockStore.Up")
			},
		},
		VersionsFunc: &StoreVersionsFunc{
			defaultHook: func(context.Context) ([]int, []int, []int, error) {
				panic("unexpected invocation of MockStore.Versions")
			},
		},
		WithMigrationLogFunc: &StoreWithMigrationLogFunc{
			defaultHook: func(context.Context, definition.Definition, bool, func() error) error {
				panic("unexpected invocation of MockStore.WithMigrationLog")
			},
		},
	}
}

// NewMockStoreFrom creates a new mock of the MockStore interface. All
// methods delegate to the given implementation, unless overwritten.
func NewMockStoreFrom(i Store) *MockStore {
	return &MockStore{
		DoneFunc: &StoreDoneFunc{
			defaultHook: i.Done,
		},
		DownFunc: &StoreDownFunc{
			defaultHook: i.Down,
		},
		IndexStatusFunc: &StoreIndexStatusFunc{
			defaultHook: i.IndexStatus,
		},
		TransactFunc: &StoreTransactFunc{
			defaultHook: i.Transact,
		},
		TryLockFunc: &StoreTryLockFunc{
			defaultHook: i.TryLock,
		},
		UpFunc: &StoreUpFunc{
			defaultHook: i.Up,
		},
		VersionsFunc: &StoreVersionsFunc{
			defaultHook: i.Versions,
		},
		WithMigrationLogFunc: &StoreWithMigrationLogFunc{
			defaultHook: i.WithMigrationLog,
		},
	}
}

// StoreDoneFunc describes the behavior when the Done method of the parent
// MockStore instance is invoked.
type StoreDoneFunc struct {
	defaultHook func(error) error
	hooks       []func(error) error
	history     []StoreDoneFuncCall
	mutex       sync.Mutex
}

// Done delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockStore) Done(v0 error) error {
	r0 := m.DoneFunc.nextHook()(v0)
	m.DoneFunc.appendCall(StoreDoneFuncCall{v0, r0})
	return r0
}

// SetDefaultHook sets function that is called when the Done method of the
// parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreDoneFunc) SetDefaultHook(hook func(error) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Done method of the parent MockStore instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *StoreDoneFunc) PushHook(hook func(error) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *StoreDoneFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(error) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *StoreDoneFunc) PushReturn(r0 error) {
	f.PushHook(func(error) error {
		return r0
	})
}

func (f *StoreDoneFunc) nextHook() func(error) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreDoneFunc) appendCall(r0 StoreDoneFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreDoneFuncCall objects describing the
// invocations of this function.
func (f *StoreDoneFunc) History() []StoreDoneFuncCall {
	f.mutex.Lock()
	history := make([]StoreDoneFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreDoneFuncCall is an object that describes an invocation of method
// Done on an instance of MockStore.
type StoreDoneFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 error
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreDoneFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreDoneFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// StoreDownFunc describes the behavior when the Down method of the parent
// MockStore instance is invoked.
type StoreDownFunc struct {
	defaultHook func(context.Context, definition.Definition) error
	hooks       []func(context.Context, definition.Definition) error
	history     []StoreDownFuncCall
	mutex       sync.Mutex
}

// Down delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockStore) Down(v0 context.Context, v1 definition.Definition) error {
	r0 := m.DownFunc.nextHook()(v0, v1)
	m.DownFunc.appendCall(StoreDownFuncCall{v0, v1, r0})
	return r0
}

// SetDefaultHook sets function that is called when the Down method of the
// parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreDownFunc) SetDefaultHook(hook func(context.Context, definition.Definition) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Down method of the parent MockStore instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *StoreDownFunc) PushHook(hook func(context.Context, definition.Definition) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *StoreDownFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, definition.Definition) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *StoreDownFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, definition.Definition) error {
		return r0
	})
}

func (f *StoreDownFunc) nextHook() func(context.Context, definition.Definition) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreDownFunc) appendCall(r0 StoreDownFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreDownFuncCall objects describing the
// invocations of this function.
func (f *StoreDownFunc) History() []StoreDownFuncCall {
	f.mutex.Lock()
	history := make([]StoreDownFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreDownFuncCall is an object that describes an invocation of method
// Down on an instance of MockStore.
type StoreDownFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 definition.Definition
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreDownFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreDownFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// StoreIndexStatusFunc describes the behavior when the IndexStatus method
// of the parent MockStore instance is invoked.
type StoreIndexStatusFunc struct {
	defaultHook func(context.Context, string, string) (storetypes.IndexStatus, bool, error)
	hooks       []func(context.Context, string, string) (storetypes.IndexStatus, bool, error)
	history     []StoreIndexStatusFuncCall
	mutex       sync.Mutex
}

// IndexStatus delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockStore) IndexStatus(v0 context.Context, v1 string, v2 string) (storetypes.IndexStatus, bool, error) {
	r0, r1, r2 := m.IndexStatusFunc.nextHook()(v0, v1, v2)
	m.IndexStatusFunc.appendCall(StoreIndexStatusFuncCall{v0, v1, v2, r0, r1, r2})
	return r0, r1, r2
}

// SetDefaultHook sets function that is called when the IndexStatus method
// of the parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreIndexStatusFunc) SetDefaultHook(hook func(context.Context, string, string) (storetypes.IndexStatus, bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// IndexStatus method of the parent MockStore instance invokes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *StoreIndexStatusFunc) PushHook(hook func(context.Context, string, string) (storetypes.IndexStatus, bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *StoreIndexStatusFunc) SetDefaultReturn(r0 storetypes.IndexStatus, r1 bool, r2 error) {
	f.SetDefaultHook(func(context.Context, string, string) (storetypes.IndexStatus, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *StoreIndexStatusFunc) PushReturn(r0 storetypes.IndexStatus, r1 bool, r2 error) {
	f.PushHook(func(context.Context, string, string) (storetypes.IndexStatus, bool, error) {
		return r0, r1, r2
	})
}

func (f *StoreIndexStatusFunc) nextHook() func(context.Context, string, string) (storetypes.IndexStatus, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreIndexStatusFunc) appendCall(r0 StoreIndexStatusFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreIndexStatusFuncCall objects describing
// the invocations of this function.
func (f *StoreIndexStatusFunc) History() []StoreIndexStatusFuncCall {
	f.mutex.Lock()
	history := make([]StoreIndexStatusFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreIndexStatusFuncCall is an object that describes an invocation of
// method IndexStatus on an instance of MockStore.
type StoreIndexStatusFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 string
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 storetypes.IndexStatus
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 bool
	// Result2 is the value of the 3rd result returned from this method
	// invocation.
	Result2 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreIndexStatusFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreIndexStatusFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1, c.Result2}
}

// StoreTransactFunc describes the behavior when the Transact method of the
// parent MockStore instance is invoked.
type StoreTransactFunc struct {
	defaultHook func(context.Context) (Store, error)
	hooks       []func(context.Context) (Store, error)
	history     []StoreTransactFuncCall
	mutex       sync.Mutex
}

// Transact delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockStore) Transact(v0 context.Context) (Store, error) {
	r0, r1 := m.TransactFunc.nextHook()(v0)
	m.TransactFunc.appendCall(StoreTransactFuncCall{v0, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the Transact method of
// the parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreTransactFunc) SetDefaultHook(hook func(context.Context) (Store, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Transact method of the parent MockStore instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *StoreTransactFunc) PushHook(hook func(context.Context) (Store, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *StoreTransactFunc) SetDefaultReturn(r0 Store, r1 error) {
	f.SetDefaultHook(func(context.Context) (Store, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *StoreTransactFunc) PushReturn(r0 Store, r1 error) {
	f.PushHook(func(context.Context) (Store, error) {
		return r0, r1
	})
}

func (f *StoreTransactFunc) nextHook() func(context.Context) (Store, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreTransactFunc) appendCall(r0 StoreTransactFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreTransactFuncCall objects describing
// the invocations of this function.
func (f *StoreTransactFunc) History() []StoreTransactFuncCall {
	f.mutex.Lock()
	history := make([]StoreTransactFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreTransactFuncCall is an object that describes an invocation of method
// Transact on an instance of MockStore.
type StoreTransactFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 Store
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreTransactFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreTransactFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// StoreTryLockFunc describes the behavior when the TryLock method of the
// parent MockStore instance is invoked.
type StoreTryLockFunc struct {
	defaultHook func(context.Context) (bool, func(err error) error, error)
	hooks       []func(context.Context) (bool, func(err error) error, error)
	history     []StoreTryLockFuncCall
	mutex       sync.Mutex
}

// TryLock delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockStore) TryLock(v0 context.Context) (bool, func(err error) error, error) {
	r0, r1, r2 := m.TryLockFunc.nextHook()(v0)
	m.TryLockFunc.appendCall(StoreTryLockFuncCall{v0, r0, r1, r2})
	return r0, r1, r2
}

// SetDefaultHook sets function that is called when the TryLock method of
// the parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreTryLockFunc) SetDefaultHook(hook func(context.Context) (bool, func(err error) error, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// TryLock method of the parent MockStore instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *StoreTryLockFunc) PushHook(hook func(context.Context) (bool, func(err error) error, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *StoreTryLockFunc) SetDefaultReturn(r0 bool, r1 func(err error) error, r2 error) {
	f.SetDefaultHook(func(context.Context) (bool, func(err error) error, error) {
		return r0, r1, r2
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *StoreTryLockFunc) PushReturn(r0 bool, r1 func(err error) error, r2 error) {
	f.PushHook(func(context.Context) (bool, func(err error) error, error) {
		return r0, r1, r2
	})
}

func (f *StoreTryLockFunc) nextHook() func(context.Context) (bool, func(err error) error, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreTryLockFunc) appendCall(r0 StoreTryLockFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreTryLockFuncCall objects describing the
// invocations of this function.
func (f *StoreTryLockFunc) History() []StoreTryLockFuncCall {
	f.mutex.Lock()
	history := make([]StoreTryLockFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreTryLockFuncCall is an object that describes an invocation of method
// TryLock on an instance of MockStore.
type StoreTryLockFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 bool
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 func(err error) error
	// Result2 is the value of the 3rd result returned from this method
	// invocation.
	Result2 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreTryLockFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreTryLockFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1, c.Result2}
}

// StoreUpFunc describes the behavior when the Up method of the parent
// MockStore instance is invoked.
type StoreUpFunc struct {
	defaultHook func(context.Context, definition.Definition) error
	hooks       []func(context.Context, definition.Definition) error
	history     []StoreUpFuncCall
	mutex       sync.Mutex
}

// Up delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockStore) Up(v0 context.Context, v1 definition.Definition) error {
	r0 := m.UpFunc.nextHook()(v0, v1)
	m.UpFunc.appendCall(StoreUpFuncCall{v0, v1, r0})
	return r0
}

// SetDefaultHook sets function that is called when the Up method of the
// parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreUpFunc) SetDefaultHook(hook func(context.Context, definition.Definition) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Up method of the parent MockStore instance invokes the hook at the front
// of the queue and discards it. After the queue is empty, the default hook
// function is invoked for any future action.
func (f *StoreUpFunc) PushHook(hook func(context.Context, definition.Definition) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *StoreUpFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, definition.Definition) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *StoreUpFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, definition.Definition) error {
		return r0
	})
}

func (f *StoreUpFunc) nextHook() func(context.Context, definition.Definition) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreUpFunc) appendCall(r0 StoreUpFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreUpFuncCall objects describing the
// invocations of this function.
func (f *StoreUpFunc) History() []StoreUpFuncCall {
	f.mutex.Lock()
	history := make([]StoreUpFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreUpFuncCall is an object that describes an invocation of method Up on
// an instance of MockStore.
type StoreUpFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 definition.Definition
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreUpFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreUpFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}

// StoreVersionsFunc describes the behavior when the Versions method of the
// parent MockStore instance is invoked.
type StoreVersionsFunc struct {
	defaultHook func(context.Context) ([]int, []int, []int, error)
	hooks       []func(context.Context) ([]int, []int, []int, error)
	history     []StoreVersionsFuncCall
	mutex       sync.Mutex
}

// Versions delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockStore) Versions(v0 context.Context) ([]int, []int, []int, error) {
	r0, r1, r2, r3 := m.VersionsFunc.nextHook()(v0)
	m.VersionsFunc.appendCall(StoreVersionsFuncCall{v0, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefaultHook sets function that is called when the Versions method of
// the parent MockStore instance is invoked and the hook queue is empty.
func (f *StoreVersionsFunc) SetDefaultHook(hook func(context.Context) ([]int, []int, []int, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Versions method of the parent MockStore instance invokes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *StoreVersionsFunc) PushHook(hook func(context.Context) ([]int, []int, []int, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *StoreVersionsFunc) SetDefaultReturn(r0 []int, r1 []int, r2 []int, r3 error) {
	f.SetDefaultHook(func(context.Context) ([]int, []int, []int, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *StoreVersionsFunc) PushReturn(r0 []int, r1 []int, r2 []int, r3 error) {
	f.PushHook(func(context.Context) ([]int, []int, []int, error) {
		return r0, r1, r2, r3
	})
}

func (f *StoreVersionsFunc) nextHook() func(context.Context) ([]int, []int, []int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreVersionsFunc) appendCall(r0 StoreVersionsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreVersionsFuncCall objects describing
// the invocations of this function.
func (f *StoreVersionsFunc) History() []StoreVersionsFuncCall {
	f.mutex.Lock()
	history := make([]StoreVersionsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreVersionsFuncCall is an object that describes an invocation of method
// Versions on an instance of MockStore.
type StoreVersionsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 []int
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 []int
	// Result2 is the value of the 3rd result returned from this method
	// invocation.
	Result2 []int
	// Result3 is the value of the 4th result returned from this method
	// invocation.
	Result3 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreVersionsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreVersionsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1, c.Result2, c.Result3}
}

// StoreWithMigrationLogFunc describes the behavior when the
// WithMigrationLog method of the parent MockStore instance is invoked.
type StoreWithMigrationLogFunc struct {
	defaultHook func(context.Context, definition.Definition, bool, func() error) error
	hooks       []func(context.Context, definition.Definition, bool, func() error) error
	history     []StoreWithMigrationLogFuncCall
	mutex       sync.Mutex
}

// WithMigrationLog delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockStore) WithMigrationLog(v0 context.Context, v1 definition.Definition, v2 bool, v3 func() error) error {
	r0 := m.WithMigrationLogFunc.nextHook()(v0, v1, v2, v3)
	m.WithMigrationLogFunc.appendCall(StoreWithMigrationLogFuncCall{v0, v1, v2, v3, r0})
	return r0
}

// SetDefaultHook sets function that is called when the WithMigrationLog
// method of the parent MockStore instance is invoked and the hook queue is
// empty.
func (f *StoreWithMigrationLogFunc) SetDefaultHook(hook func(context.Context, definition.Definition, bool, func() error) error) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// WithMigrationLog method of the parent MockStore instance invokes the hook
// at the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *StoreWithMigrationLogFunc) PushHook(hook func(context.Context, definition.Definition, bool, func() error) error) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *StoreWithMigrationLogFunc) SetDefaultReturn(r0 error) {
	f.SetDefaultHook(func(context.Context, definition.Definition, bool, func() error) error {
		return r0
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *StoreWithMigrationLogFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, definition.Definition, bool, func() error) error {
		return r0
	})
}

func (f *StoreWithMigrationLogFunc) nextHook() func(context.Context, definition.Definition, bool, func() error) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreWithMigrationLogFunc) appendCall(r0 StoreWithMigrationLogFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of StoreWithMigrationLogFuncCall objects
// describing the invocations of this function.
func (f *StoreWithMigrationLogFunc) History() []StoreWithMigrationLogFuncCall {
	f.mutex.Lock()
	history := make([]StoreWithMigrationLogFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreWithMigrationLogFuncCall is an object that describes an invocation
// of method WithMigrationLog on an instance of MockStore.
type StoreWithMigrationLogFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 definition.Definition
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 bool
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 func() error
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c StoreWithMigrationLogFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c StoreWithMigrationLogFuncCall) Results() []interface{} {
	return []interface{}{c.Result0}
}
