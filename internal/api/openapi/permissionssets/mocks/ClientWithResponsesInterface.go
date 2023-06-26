// Code generated by mockery v2.30.1. DO NOT EDIT.

package mocks

import (
	context "context"

	client "github.com/stax-labs/terraform-provider-stax/internal/api/openapi/permissionssets/client"

	io "io"

	mock "github.com/stretchr/testify/mock"

	models "github.com/stax-labs/terraform-provider-stax/internal/api/openapi/permissionssets/models"

	uuid "github.com/google/uuid"
)

// ClientWithResponsesInterface is an autogenerated mock type for the ClientWithResponsesInterface type
type ClientWithResponsesInterface struct {
	mock.Mock
}

// CreatePermissionSetAssignmentsWithBodyWithResponse provides a mock function with given fields: ctx, permissionSetId, contentType, body, reqEditors
func (_m *ClientWithResponsesInterface) CreatePermissionSetAssignmentsWithBodyWithResponse(ctx context.Context, permissionSetId uuid.UUID, contentType string, body io.Reader, reqEditors ...client.RequestEditorFn) (*client.CreatePermissionSetAssignmentsResponse, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, permissionSetId, contentType, body)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *client.CreatePermissionSetAssignmentsResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string, io.Reader, ...client.RequestEditorFn) (*client.CreatePermissionSetAssignmentsResponse, error)); ok {
		return rf(ctx, permissionSetId, contentType, body, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string, io.Reader, ...client.RequestEditorFn) *client.CreatePermissionSetAssignmentsResponse); ok {
		r0 = rf(ctx, permissionSetId, contentType, body, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.CreatePermissionSetAssignmentsResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, string, io.Reader, ...client.RequestEditorFn) error); ok {
		r1 = rf(ctx, permissionSetId, contentType, body, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreatePermissionSetAssignmentsWithResponse provides a mock function with given fields: ctx, permissionSetId, body, reqEditors
func (_m *ClientWithResponsesInterface) CreatePermissionSetAssignmentsWithResponse(ctx context.Context, permissionSetId uuid.UUID, body []struct {
	AccountTypeId uuid.UUID `json:"AccountTypeId"`
	GroupId       uuid.UUID `json:"GroupId"`
}, reqEditors ...client.RequestEditorFn) (*client.CreatePermissionSetAssignmentsResponse, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, permissionSetId, body)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *client.CreatePermissionSetAssignmentsResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, []struct {
		AccountTypeId uuid.UUID `json:"AccountTypeId"`
		GroupId       uuid.UUID `json:"GroupId"`
	}, ...client.RequestEditorFn) (*client.CreatePermissionSetAssignmentsResponse, error)); ok {
		return rf(ctx, permissionSetId, body, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, []struct {
		AccountTypeId uuid.UUID `json:"AccountTypeId"`
		GroupId       uuid.UUID `json:"GroupId"`
	}, ...client.RequestEditorFn) *client.CreatePermissionSetAssignmentsResponse); ok {
		r0 = rf(ctx, permissionSetId, body, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.CreatePermissionSetAssignmentsResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, []struct {
		AccountTypeId uuid.UUID `json:"AccountTypeId"`
		GroupId       uuid.UUID `json:"GroupId"`
	}, ...client.RequestEditorFn) error); ok {
		r1 = rf(ctx, permissionSetId, body, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreatePermissionSetWithBodyWithResponse provides a mock function with given fields: ctx, contentType, body, reqEditors
func (_m *ClientWithResponsesInterface) CreatePermissionSetWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...client.RequestEditorFn) (*client.CreatePermissionSetResponse, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, contentType, body)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *client.CreatePermissionSetResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, io.Reader, ...client.RequestEditorFn) (*client.CreatePermissionSetResponse, error)); ok {
		return rf(ctx, contentType, body, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, io.Reader, ...client.RequestEditorFn) *client.CreatePermissionSetResponse); ok {
		r0 = rf(ctx, contentType, body, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.CreatePermissionSetResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, io.Reader, ...client.RequestEditorFn) error); ok {
		r1 = rf(ctx, contentType, body, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreatePermissionSetWithResponse provides a mock function with given fields: ctx, body, reqEditors
func (_m *ClientWithResponsesInterface) CreatePermissionSetWithResponse(ctx context.Context, body models.CreatePermissionSetRecord, reqEditors ...client.RequestEditorFn) (*client.CreatePermissionSetResponse, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, body)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *client.CreatePermissionSetResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, models.CreatePermissionSetRecord, ...client.RequestEditorFn) (*client.CreatePermissionSetResponse, error)); ok {
		return rf(ctx, body, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, models.CreatePermissionSetRecord, ...client.RequestEditorFn) *client.CreatePermissionSetResponse); ok {
		r0 = rf(ctx, body, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.CreatePermissionSetResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, models.CreatePermissionSetRecord, ...client.RequestEditorFn) error); ok {
		r1 = rf(ctx, body, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeletePermissionSetAssignmentWithResponse provides a mock function with given fields: ctx, permissionSetId, assignmentId, reqEditors
func (_m *ClientWithResponsesInterface) DeletePermissionSetAssignmentWithResponse(ctx context.Context, permissionSetId uuid.UUID, assignmentId uuid.UUID, reqEditors ...client.RequestEditorFn) (*client.DeletePermissionSetAssignmentResponse, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, permissionSetId, assignmentId)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *client.DeletePermissionSetAssignmentResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, uuid.UUID, ...client.RequestEditorFn) (*client.DeletePermissionSetAssignmentResponse, error)); ok {
		return rf(ctx, permissionSetId, assignmentId, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, uuid.UUID, ...client.RequestEditorFn) *client.DeletePermissionSetAssignmentResponse); ok {
		r0 = rf(ctx, permissionSetId, assignmentId, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.DeletePermissionSetAssignmentResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, uuid.UUID, ...client.RequestEditorFn) error); ok {
		r1 = rf(ctx, permissionSetId, assignmentId, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeletePermissionSetWithResponse provides a mock function with given fields: ctx, permissionSetId, reqEditors
func (_m *ClientWithResponsesInterface) DeletePermissionSetWithResponse(ctx context.Context, permissionSetId uuid.UUID, reqEditors ...client.RequestEditorFn) (*client.DeletePermissionSetResponse, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, permissionSetId)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *client.DeletePermissionSetResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, ...client.RequestEditorFn) (*client.DeletePermissionSetResponse, error)); ok {
		return rf(ctx, permissionSetId, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, ...client.RequestEditorFn) *client.DeletePermissionSetResponse); ok {
		r0 = rf(ctx, permissionSetId, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.DeletePermissionSetResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, ...client.RequestEditorFn) error); ok {
		r1 = rf(ctx, permissionSetId, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAPIDocumentWithResponse provides a mock function with given fields: ctx, reqEditors
func (_m *ClientWithResponsesInterface) GetAPIDocumentWithResponse(ctx context.Context, reqEditors ...client.RequestEditorFn) (*client.GetAPIDocumentResponse, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *client.GetAPIDocumentResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, ...client.RequestEditorFn) (*client.GetAPIDocumentResponse, error)); ok {
		return rf(ctx, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, ...client.RequestEditorFn) *client.GetAPIDocumentResponse); ok {
		r0 = rf(ctx, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.GetAPIDocumentResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, ...client.RequestEditorFn) error); ok {
		r1 = rf(ctx, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPermissionSetWithResponse provides a mock function with given fields: ctx, permissionSetId, reqEditors
func (_m *ClientWithResponsesInterface) GetPermissionSetWithResponse(ctx context.Context, permissionSetId uuid.UUID, reqEditors ...client.RequestEditorFn) (*client.GetPermissionSetResponse, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, permissionSetId)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *client.GetPermissionSetResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, ...client.RequestEditorFn) (*client.GetPermissionSetResponse, error)); ok {
		return rf(ctx, permissionSetId, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, ...client.RequestEditorFn) *client.GetPermissionSetResponse); ok {
		r0 = rf(ctx, permissionSetId, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.GetPermissionSetResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, ...client.RequestEditorFn) error); ok {
		r1 = rf(ctx, permissionSetId, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAWSManagedPoliciesWithResponse provides a mock function with given fields: ctx, reqEditors
func (_m *ClientWithResponsesInterface) ListAWSManagedPoliciesWithResponse(ctx context.Context, reqEditors ...client.RequestEditorFn) (*client.ListAWSManagedPoliciesResponse, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *client.ListAWSManagedPoliciesResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, ...client.RequestEditorFn) (*client.ListAWSManagedPoliciesResponse, error)); ok {
		return rf(ctx, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, ...client.RequestEditorFn) *client.ListAWSManagedPoliciesResponse); ok {
		r0 = rf(ctx, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.ListAWSManagedPoliciesResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, ...client.RequestEditorFn) error); ok {
		r1 = rf(ctx, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListMyRolesWithResponse provides a mock function with given fields: ctx, params, reqEditors
func (_m *ClientWithResponsesInterface) ListMyRolesWithResponse(ctx context.Context, params *models.ListMyRolesParams, reqEditors ...client.RequestEditorFn) (*client.ListMyRolesResponse, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, params)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *client.ListMyRolesResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.ListMyRolesParams, ...client.RequestEditorFn) (*client.ListMyRolesResponse, error)); ok {
		return rf(ctx, params, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *models.ListMyRolesParams, ...client.RequestEditorFn) *client.ListMyRolesResponse); ok {
		r0 = rf(ctx, params, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.ListMyRolesResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *models.ListMyRolesParams, ...client.RequestEditorFn) error); ok {
		r1 = rf(ctx, params, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListPermissionSetAssignmentsWithResponse provides a mock function with given fields: ctx, permissionSetId, params, reqEditors
func (_m *ClientWithResponsesInterface) ListPermissionSetAssignmentsWithResponse(ctx context.Context, permissionSetId uuid.UUID, params *models.ListPermissionSetAssignmentsParams, reqEditors ...client.RequestEditorFn) (*client.ListPermissionSetAssignmentsResponse, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, permissionSetId, params)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *client.ListPermissionSetAssignmentsResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *models.ListPermissionSetAssignmentsParams, ...client.RequestEditorFn) (*client.ListPermissionSetAssignmentsResponse, error)); ok {
		return rf(ctx, permissionSetId, params, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *models.ListPermissionSetAssignmentsParams, ...client.RequestEditorFn) *client.ListPermissionSetAssignmentsResponse); ok {
		r0 = rf(ctx, permissionSetId, params, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.ListPermissionSetAssignmentsResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, *models.ListPermissionSetAssignmentsParams, ...client.RequestEditorFn) error); ok {
		r1 = rf(ctx, permissionSetId, params, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListPermissionSetsWithResponse provides a mock function with given fields: ctx, params, reqEditors
func (_m *ClientWithResponsesInterface) ListPermissionSetsWithResponse(ctx context.Context, params *models.ListPermissionSetsParams, reqEditors ...client.RequestEditorFn) (*client.ListPermissionSetsResponse, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, params)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *client.ListPermissionSetsResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.ListPermissionSetsParams, ...client.RequestEditorFn) (*client.ListPermissionSetsResponse, error)); ok {
		return rf(ctx, params, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *models.ListPermissionSetsParams, ...client.RequestEditorFn) *client.ListPermissionSetsResponse); ok {
		r0 = rf(ctx, params, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.ListPermissionSetsResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *models.ListPermissionSetsParams, ...client.RequestEditorFn) error); ok {
		r1 = rf(ctx, params, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListRolesWithResponse provides a mock function with given fields: ctx, params, reqEditors
func (_m *ClientWithResponsesInterface) ListRolesWithResponse(ctx context.Context, params *models.ListRolesParams, reqEditors ...client.RequestEditorFn) (*client.ListRolesResponse, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, params)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *client.ListRolesResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.ListRolesParams, ...client.RequestEditorFn) (*client.ListRolesResponse, error)); ok {
		return rf(ctx, params, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *models.ListRolesParams, ...client.RequestEditorFn) *client.ListRolesResponse); ok {
		r0 = rf(ctx, params, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.ListRolesResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *models.ListRolesParams, ...client.RequestEditorFn) error); ok {
		r1 = rf(ctx, params, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RedeployPermissionSetAssignmentWithResponse provides a mock function with given fields: ctx, permissionSetId, assignmentId, reqEditors
func (_m *ClientWithResponsesInterface) RedeployPermissionSetAssignmentWithResponse(ctx context.Context, permissionSetId uuid.UUID, assignmentId uuid.UUID, reqEditors ...client.RequestEditorFn) (*client.RedeployPermissionSetAssignmentResponse, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, permissionSetId, assignmentId)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *client.RedeployPermissionSetAssignmentResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, uuid.UUID, ...client.RequestEditorFn) (*client.RedeployPermissionSetAssignmentResponse, error)); ok {
		return rf(ctx, permissionSetId, assignmentId, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, uuid.UUID, ...client.RequestEditorFn) *client.RedeployPermissionSetAssignmentResponse); ok {
		r0 = rf(ctx, permissionSetId, assignmentId, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.RedeployPermissionSetAssignmentResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, uuid.UUID, ...client.RequestEditorFn) error); ok {
		r1 = rf(ctx, permissionSetId, assignmentId, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdatePermissionSetWithBodyWithResponse provides a mock function with given fields: ctx, permissionSetId, contentType, body, reqEditors
func (_m *ClientWithResponsesInterface) UpdatePermissionSetWithBodyWithResponse(ctx context.Context, permissionSetId uuid.UUID, contentType string, body io.Reader, reqEditors ...client.RequestEditorFn) (*client.UpdatePermissionSetResponse, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, permissionSetId, contentType, body)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *client.UpdatePermissionSetResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string, io.Reader, ...client.RequestEditorFn) (*client.UpdatePermissionSetResponse, error)); ok {
		return rf(ctx, permissionSetId, contentType, body, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string, io.Reader, ...client.RequestEditorFn) *client.UpdatePermissionSetResponse); ok {
		r0 = rf(ctx, permissionSetId, contentType, body, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.UpdatePermissionSetResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, string, io.Reader, ...client.RequestEditorFn) error); ok {
		r1 = rf(ctx, permissionSetId, contentType, body, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdatePermissionSetWithResponse provides a mock function with given fields: ctx, permissionSetId, body, reqEditors
func (_m *ClientWithResponsesInterface) UpdatePermissionSetWithResponse(ctx context.Context, permissionSetId uuid.UUID, body models.UpdatePermissionSetRecord, reqEditors ...client.RequestEditorFn) (*client.UpdatePermissionSetResponse, error) {
	_va := make([]interface{}, len(reqEditors))
	for _i := range reqEditors {
		_va[_i] = reqEditors[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, permissionSetId, body)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *client.UpdatePermissionSetResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, models.UpdatePermissionSetRecord, ...client.RequestEditorFn) (*client.UpdatePermissionSetResponse, error)); ok {
		return rf(ctx, permissionSetId, body, reqEditors...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, models.UpdatePermissionSetRecord, ...client.RequestEditorFn) *client.UpdatePermissionSetResponse); ok {
		r0 = rf(ctx, permissionSetId, body, reqEditors...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.UpdatePermissionSetResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, models.UpdatePermissionSetRecord, ...client.RequestEditorFn) error); ok {
		r1 = rf(ctx, permissionSetId, body, reqEditors...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewClientWithResponsesInterface creates a new instance of ClientWithResponsesInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewClientWithResponsesInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *ClientWithResponsesInterface {
	mock := &ClientWithResponsesInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}