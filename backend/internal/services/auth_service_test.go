package services

import "testing"

func Test_generateToken(t *testing.T) {
	type args struct {
		userID   int64
		userType int8
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateToken(tt.args.userID, tt.args.userType)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("generateToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}
