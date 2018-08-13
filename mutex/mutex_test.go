package mutex

import (
	"reflect"
	"testing"
)

func TestNewMutex(t *testing.T) {
	type args struct {
		lockCap int
	}
	tests := []struct {
		name    string
		args    args
		want    *Mutex
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "0 lockcap",
			args: args{
				lockCap: 0,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMutex(tt.args.lockCap)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMutex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMutex() = %v, want %v", got, tt.want)
			}
		})
	}
}
