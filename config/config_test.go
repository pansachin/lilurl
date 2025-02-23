// Config package deals with applciation configuration
package config

import "testing"

func TestRead(t *testing.T) {
	type args struct {
		cfg   interface{}
		paths []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Read(tt.args.cfg, tt.args.paths...); (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
