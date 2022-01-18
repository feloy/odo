package catalog

import (
	"reflect"
	"testing"
)

func TestDevfileComponentTypeList_GetLanguages(t *testing.T) {
	type fields struct {
		Items []DevfileComponentType
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "no devfiles",
			want: []string{},
		},
		{
			name: "some devfiles",
			fields: fields{
				Items: []DevfileComponentType{
					{
						Name:        "devfile4",
						DisplayName: "first devfile for lang3",
						Registry: Registry{
							Name: "Registry1",
						},
						Language: "lang3",
					},
					{
						Name:        "devfile1",
						DisplayName: "first devfile for lang1",
						Registry: Registry{
							Name: "Registry2",
						},
						Language: "lang1",
					},
					{
						Name:        "devfile3",
						DisplayName: "another devfile for lang2",
						Registry: Registry{
							Name: "Registry1",
						},
						Language: "lang2",
					},
					{
						Name:        "devfile2",
						DisplayName: "second devfile for lang1",
						Registry: Registry{
							Name: "Registry1",
						},
						Language: "lang1",
					},
				},
			},
			want: []string{"lang1", "lang2", "lang3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &DevfileComponentTypeList{
				Items: tt.fields.Items,
			}
			if got := o.GetLanguages(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DevfileComponentTypeList.GetLanguages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDevfileComponentTypeList_GetProjectTypes(t *testing.T) {
	type fields struct {
		Items []DevfileComponentType
	}
	type args struct {
		language string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   TypesWithDetails
	}{
		{
			name: "No devfiles => no project types",
			want: TypesWithDetails{},
		},
		{
			name: "project types for lang1",
			fields: fields{
				Items: []DevfileComponentType{
					{
						Name:        "devfile4",
						DisplayName: "first devfile for lang3",
						Registry: Registry{
							Name: "Registry1",
						},
						Language: "lang3",
					},
					{
						Name:        "devfile1",
						DisplayName: "first devfile for lang1",
						Registry: Registry{
							Name: "Registry1",
						},
						Language: "lang1",
					},
					{
						Name:        "devfile1",
						DisplayName: "first devfile for lang1",
						Registry: Registry{
							Name: "Registry2",
						},
						Language: "lang1",
					},
					{
						Name:        "devfile3",
						DisplayName: "another devfile for lang2",
						Registry: Registry{
							Name: "Registry1",
						},
						Language: "lang2",
					},
					{
						Name:        "devfile2",
						DisplayName: "second devfile for lang1",
						Registry: Registry{
							Name: "Registry1",
						},
						Language: "lang1",
					},
				},
			},
			args: args{
				language: "lang1",
			},
			want: TypesWithDetails{
				"first devfile for lang1": []DevfileComponentType{
					{
						Name:        "devfile1",
						DisplayName: "first devfile for lang1",
						Language:    "lang1",
						Registry: Registry{
							Name: "Registry1",
						},
					},
					{
						Name:        "devfile1",
						DisplayName: "first devfile for lang1",
						Language:    "lang1",
						Registry: Registry{
							Name: "Registry2",
						},
					},
				},
				"second devfile for lang1": []DevfileComponentType{
					{
						Name:        "devfile2",
						DisplayName: "second devfile for lang1",
						Language:    "lang1",
						Registry: Registry{
							Name: "Registry1",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &DevfileComponentTypeList{
				Items: tt.fields.Items,
			}
			if got := o.GetProjectTypes(tt.args.language); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DevfileComponentTypeList.GetProjectTypes() = \n%+v, want \n%+v", got, tt.want)
			}
		})
	}
}
