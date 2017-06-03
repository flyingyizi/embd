package udoo

import (
	"reflect"
	"testing"
)

func TestGetBoardInfo(t *testing.T) {
	tests := []struct {
		name     string
		wantInfo InfoT
		wantErr  bool
	}{
		// TODO: Add test cases.
		{name: "in Neo extend", wantInfo: InfoT{Model: ModelNeoExtend, HasM4: true, HasLvds15: true}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInfo, err := GetBoardInfo()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBoardInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotInfo, tt.wantInfo) {
				t.Errorf("GetBoardInfo() = %v, want %v", gotInfo, tt.wantInfo)
			}
		})
	}
}
