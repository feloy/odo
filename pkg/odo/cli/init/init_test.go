package init

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

func TestInitOptions_Complete(t *testing.T) {
	type fields struct {
		backends func(*gomock.Controller) []ParamsBuilder
	}
	tests := []struct {
		name           string
		fields         fields
		cmdlineExpects func(*cmdline.MockCmdline)
		fsysPopulate   func(fsys filesystem.Filesystem)
		wantErr        bool
	}{
		{
			name: "directory not empty",
			fsysPopulate: func(fsys filesystem.Filesystem) {
				_ = fsys.WriteFile(".emptyfile", []byte(""), 0644)
			},
			wantErr: true,
		},
		{
			name: "second backend used",
			fields: fields{
				backends: func(ctrl *gomock.Controller) []ParamsBuilder {
					b1 := NewMockParamsBuilder(ctrl)
					b2 := NewMockParamsBuilder(ctrl)
					b1.EXPECT().IsAdequate(gomock.Any()).Return(false)
					b2.EXPECT().IsAdequate(gomock.Any()).Return(true)
					b2.EXPECT().ParamsBuild().Times(1)
					return []ParamsBuilder{b1, b2}
				},
			},
			cmdlineExpects: func(mock *cmdline.MockCmdline) {
				mock.EXPECT().GetFlags()
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := filesystem.NewFakeFs()
			if tt.fsysPopulate != nil {
				tt.fsysPopulate(fsys)
			}
			ctrl := gomock.NewController(t)
			var backends []ParamsBuilder
			if tt.fields.backends != nil {
				backends = tt.fields.backends(ctrl)
			}
			o := NewInitOptions(backends, fsys)
			cmdline := cmdline.NewMockCmdline(ctrl)
			if tt.cmdlineExpects != nil {
				tt.cmdlineExpects(cmdline)
			}
			if err := o.Complete(cmdline, []string{}); (err != nil) != tt.wantErr {
				t.Errorf("InitOptions.Complete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
