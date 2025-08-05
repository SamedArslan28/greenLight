package main

import "testing"

func TestOpenDB(t *testing.T) {
	LoadEnvOnce("/Users/abdulsamedarslan/Desktop/greenLight/.env")
	dsn := mustGetEnv("DB_DSN")
	tests := []struct {
		name    string
		config  config
		wantErr bool
	}{
		{
			name: "Valid DSN",
			config: config{
				db: dbConfig{
					dsn:          dsn,
					maxOpenConns: 10,
					maxIdleConns: 10,
					maxIdleTime:  "1m",
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid DSN",
			config: config{
				db: dbConfig{
					dsn:          "invalid-dsn",
					maxOpenConns: 10,
					maxIdleConns: 10,
					maxIdleTime:  "1m",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := openDB(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("openDB() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
