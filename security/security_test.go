package security

import (
	"testing"

	"rest-authentication/config"
	"rest-authentication/storage"
)

func TestSecurity_GenerateTokens(t *testing.T) {
	type fields struct {
		cfg     config.Security
		storage *storage.Storage
	}
	type args struct {
		GUID string
		ip   string
		UUID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Test 1",
			fields: fields{
				cfg: config.Security{
					SecretKey: "secret",
				},
			},
			args: args{
				GUID: "guid",
				ip:   "ip",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Security{
				cfg:     tt.fields.cfg,
				storage: tt.fields.storage,
			}
			access, err := s.generateAccessToken(tt.args.GUID, tt.args.ip, tt.args.UUID)
			if err != nil {
				t.Errorf("GenerateTokens() error = %v", err)
				return
			}
			refresh, err := s.generateRefreshToken(access)
			if err != nil {
				t.Errorf("GenerateTokens() error = %v", err)
				return
			}
			if _, err = s.validateToken(access); err != nil {
				t.Errorf("access token is invalid")
			}
			if !s.ValidateRefresh(access, refresh) {
				t.Errorf("refresh token is invalid")
			}
		})
	}
}
