package dbmodels

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithCreatedOn_Validate(t *testing.T) {
	type fields struct {
		CreatedOn string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "ok",
			fields: fields{
				CreatedOn: "2020-12-31",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err, i...)
			},
		},
		{
			name: "missing",
			fields: fields{
				CreatedOn: "",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		},
		{
			name: "invalid_letters",
			fields: fields{
				CreatedOn: "abc",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		},
		{
			name: "invalid_no_dashes",
			fields: fields{
				CreatedOn: "20201231",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		},
		{
			name: "invalid_slashes",
			fields: fields{
				CreatedOn: "2020/12/31",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		},
		{
			name: "invalid_us",
			fields: fields{
				CreatedOn: "31/12/2020",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &WithCreatedOn{
				CreatedOn: tt.fields.CreatedOn,
			}
			tt.wantErr(t, v.Validate(), fmt.Sprintf("{CreatedOn=%s}.Validate()", v.CreatedOn))
		})
	}
}
