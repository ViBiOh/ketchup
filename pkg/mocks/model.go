// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ViBiOh/ketchup/pkg/model (interfaces: Mailer,AuthService,UserService,UserStore,HelmProvider,RepositoryService)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	model "github.com/ViBiOh/auth/v2/pkg/model"
	model0 "github.com/ViBiOh/ketchup/pkg/model"
	semver "github.com/ViBiOh/ketchup/pkg/semver"
	model1 "github.com/ViBiOh/mailer/pkg/model"
	gomock "github.com/golang/mock/gomock"
)

// Mailer is a mock of Mailer interface.
type Mailer struct {
	ctrl     *gomock.Controller
	recorder *MailerMockRecorder
}

// MailerMockRecorder is the mock recorder for Mailer.
type MailerMockRecorder struct {
	mock *Mailer
}

// NewMailer creates a new mock instance.
func NewMailer(ctrl *gomock.Controller) *Mailer {
	mock := &Mailer{ctrl: ctrl}
	mock.recorder = &MailerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mailer) EXPECT() *MailerMockRecorder {
	return m.recorder
}

// Enabled mocks base method.
func (m *Mailer) Enabled() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Enabled")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Enabled indicates an expected call of Enabled.
func (mr *MailerMockRecorder) Enabled() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Enabled", reflect.TypeOf((*Mailer)(nil).Enabled))
}

// Send mocks base method.
func (m *Mailer) Send(arg0 context.Context, arg1 model1.MailRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MailerMockRecorder) Send(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*Mailer)(nil).Send), arg0, arg1)
}

// AuthService is a mock of AuthService interface.
type AuthService struct {
	ctrl     *gomock.Controller
	recorder *AuthServiceMockRecorder
}

// AuthServiceMockRecorder is the mock recorder for AuthService.
type AuthServiceMockRecorder struct {
	mock *AuthService
}

// NewAuthService creates a new mock instance.
func NewAuthService(ctrl *gomock.Controller) *AuthService {
	mock := &AuthService{ctrl: ctrl}
	mock.recorder = &AuthServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *AuthService) EXPECT() *AuthServiceMockRecorder {
	return m.recorder
}

// Check mocks base method.
func (m *AuthService) Check(arg0 context.Context, arg1, arg2 model.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Check", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Check indicates an expected call of Check.
func (mr *AuthServiceMockRecorder) Check(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Check", reflect.TypeOf((*AuthService)(nil).Check), arg0, arg1, arg2)
}

// Create mocks base method.
func (m *AuthService) Create(arg0 context.Context, arg1 model.User) (model.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1)
	ret0, _ := ret[0].(model.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *AuthServiceMockRecorder) Create(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*AuthService)(nil).Create), arg0, arg1)
}

// UserService is a mock of UserService interface.
type UserService struct {
	ctrl     *gomock.Controller
	recorder *UserServiceMockRecorder
}

// UserServiceMockRecorder is the mock recorder for UserService.
type UserServiceMockRecorder struct {
	mock *UserService
}

// NewUserService creates a new mock instance.
func NewUserService(ctrl *gomock.Controller) *UserService {
	mock := &UserService{ctrl: ctrl}
	mock.recorder = &UserServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *UserService) EXPECT() *UserServiceMockRecorder {
	return m.recorder
}

// StoreInContext mocks base method.
func (m *UserService) StoreInContext(arg0 context.Context) context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreInContext", arg0)
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// StoreInContext indicates an expected call of StoreInContext.
func (mr *UserServiceMockRecorder) StoreInContext(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreInContext", reflect.TypeOf((*UserService)(nil).StoreInContext), arg0)
}

// UserStore is a mock of UserStore interface.
type UserStore struct {
	ctrl     *gomock.Controller
	recorder *UserStoreMockRecorder
}

// UserStoreMockRecorder is the mock recorder for UserStore.
type UserStoreMockRecorder struct {
	mock *UserStore
}

// NewUserStore creates a new mock instance.
func NewUserStore(ctrl *gomock.Controller) *UserStore {
	mock := &UserStore{ctrl: ctrl}
	mock.recorder = &UserStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *UserStore) EXPECT() *UserStoreMockRecorder {
	return m.recorder
}

// Count mocks base method.
func (m *UserStore) Count(arg0 context.Context) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Count", arg0)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Count indicates an expected call of Count.
func (mr *UserStoreMockRecorder) Count(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Count", reflect.TypeOf((*UserStore)(nil).Count), arg0)
}

// Create mocks base method.
func (m *UserStore) Create(arg0 context.Context, arg1 model0.User) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *UserStoreMockRecorder) Create(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*UserStore)(nil).Create), arg0, arg1)
}

// DoAtomic mocks base method.
func (m *UserStore) DoAtomic(arg0 context.Context, arg1 func(context.Context) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DoAtomic", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DoAtomic indicates an expected call of DoAtomic.
func (mr *UserStoreMockRecorder) DoAtomic(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DoAtomic", reflect.TypeOf((*UserStore)(nil).DoAtomic), arg0, arg1)
}

// GetByEmail mocks base method.
func (m *UserStore) GetByEmail(arg0 context.Context, arg1 string) (model0.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByEmail", arg0, arg1)
	ret0, _ := ret[0].(model0.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByEmail indicates an expected call of GetByEmail.
func (mr *UserStoreMockRecorder) GetByEmail(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByEmail", reflect.TypeOf((*UserStore)(nil).GetByEmail), arg0, arg1)
}

// GetByLoginID mocks base method.
func (m *UserStore) GetByLoginID(arg0 context.Context, arg1 uint64) (model0.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByLoginID", arg0, arg1)
	ret0, _ := ret[0].(model0.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByLoginID indicates an expected call of GetByLoginID.
func (mr *UserStoreMockRecorder) GetByLoginID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByLoginID", reflect.TypeOf((*UserStore)(nil).GetByLoginID), arg0, arg1)
}

// HelmProvider is a mock of HelmProvider interface.
type HelmProvider struct {
	ctrl     *gomock.Controller
	recorder *HelmProviderMockRecorder
}

// HelmProviderMockRecorder is the mock recorder for HelmProvider.
type HelmProviderMockRecorder struct {
	mock *HelmProvider
}

// NewHelmProvider creates a new mock instance.
func NewHelmProvider(ctrl *gomock.Controller) *HelmProvider {
	mock := &HelmProvider{ctrl: ctrl}
	mock.recorder = &HelmProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *HelmProvider) EXPECT() *HelmProviderMockRecorder {
	return m.recorder
}

// FetchIndex mocks base method.
func (m *HelmProvider) FetchIndex(arg0 string, arg1 map[string][]string) (map[string]map[string]semver.Version, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchIndex", arg0, arg1)
	ret0, _ := ret[0].(map[string]map[string]semver.Version)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchIndex indicates an expected call of FetchIndex.
func (mr *HelmProviderMockRecorder) FetchIndex(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchIndex", reflect.TypeOf((*HelmProvider)(nil).FetchIndex), arg0, arg1)
}

// LatestVersions mocks base method.
func (m *HelmProvider) LatestVersions(arg0, arg1 string, arg2 []string) (map[string]semver.Version, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LatestVersions", arg0, arg1, arg2)
	ret0, _ := ret[0].(map[string]semver.Version)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LatestVersions indicates an expected call of LatestVersions.
func (mr *HelmProviderMockRecorder) LatestVersions(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LatestVersions", reflect.TypeOf((*HelmProvider)(nil).LatestVersions), arg0, arg1, arg2)
}

// RepositoryService is a mock of RepositoryService interface.
type RepositoryService struct {
	ctrl     *gomock.Controller
	recorder *RepositoryServiceMockRecorder
}

// RepositoryServiceMockRecorder is the mock recorder for RepositoryService.
type RepositoryServiceMockRecorder struct {
	mock *RepositoryService
}

// NewRepositoryService creates a new mock instance.
func NewRepositoryService(ctrl *gomock.Controller) *RepositoryService {
	mock := &RepositoryService{ctrl: ctrl}
	mock.recorder = &RepositoryServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *RepositoryService) EXPECT() *RepositoryServiceMockRecorder {
	return m.recorder
}

// Clean mocks base method.
func (m *RepositoryService) Clean(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Clean", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Clean indicates an expected call of Clean.
func (mr *RepositoryServiceMockRecorder) Clean(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Clean", reflect.TypeOf((*RepositoryService)(nil).Clean), arg0)
}

// GetOrCreate mocks base method.
func (m *RepositoryService) GetOrCreate(arg0 context.Context, arg1 model0.RepositoryKind, arg2, arg3, arg4 string) (model0.Repository, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrCreate", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(model0.Repository)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrCreate indicates an expected call of GetOrCreate.
func (mr *RepositoryServiceMockRecorder) GetOrCreate(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrCreate", reflect.TypeOf((*RepositoryService)(nil).GetOrCreate), arg0, arg1, arg2, arg3, arg4)
}

// LatestVersions mocks base method.
func (m *RepositoryService) LatestVersions(arg0 model0.Repository) (map[string]semver.Version, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LatestVersions", arg0)
	ret0, _ := ret[0].(map[string]semver.Version)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LatestVersions indicates an expected call of LatestVersions.
func (mr *RepositoryServiceMockRecorder) LatestVersions(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LatestVersions", reflect.TypeOf((*RepositoryService)(nil).LatestVersions), arg0)
}

// List mocks base method.
func (m *RepositoryService) List(arg0 context.Context, arg1 uint, arg2 string) ([]model0.Repository, uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", arg0, arg1, arg2)
	ret0, _ := ret[0].([]model0.Repository)
	ret1, _ := ret[1].(uint64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// List indicates an expected call of List.
func (mr *RepositoryServiceMockRecorder) List(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*RepositoryService)(nil).List), arg0, arg1, arg2)
}

// ListByKinds mocks base method.
func (m *RepositoryService) ListByKinds(arg0 context.Context, arg1 uint, arg2 string, arg3 ...model0.RepositoryKind) ([]model0.Repository, uint64, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListByKinds", varargs...)
	ret0, _ := ret[0].([]model0.Repository)
	ret1, _ := ret[1].(uint64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ListByKinds indicates an expected call of ListByKinds.
func (mr *RepositoryServiceMockRecorder) ListByKinds(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByKinds", reflect.TypeOf((*RepositoryService)(nil).ListByKinds), varargs...)
}

// Suggest mocks base method.
func (m *RepositoryService) Suggest(arg0 context.Context, arg1 []uint64, arg2 uint64) ([]model0.Repository, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Suggest", arg0, arg1, arg2)
	ret0, _ := ret[0].([]model0.Repository)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Suggest indicates an expected call of Suggest.
func (mr *RepositoryServiceMockRecorder) Suggest(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Suggest", reflect.TypeOf((*RepositoryService)(nil).Suggest), arg0, arg1, arg2)
}

// Update mocks base method.
func (m *RepositoryService) Update(arg0 context.Context, arg1 model0.Repository) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *RepositoryServiceMockRecorder) Update(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*RepositoryService)(nil).Update), arg0, arg1)
}