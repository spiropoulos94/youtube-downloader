package validators

import (
	"strings"
	"testing"
)

func TestYouTubeURLValidator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid YouTube URL",
			url:     "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			wantErr: false,
		},
		{
			name:    "Empty URL",
			url:     "",
			wantErr: true,
			errMsg:  "URL cannot be empty",
		},
		{
			name:    "Malformed URL",
			url:     "http://[::1]:namedport",
			wantErr: true,
			errMsg:  "invalid URL format",
		},
		{
			name:    "Non-YouTube domain",
			url:     "not-a-url",
			wantErr: true,
			errMsg:  "not a YouTube URL",
		},
		{
			name:    "Non-YouTube URL",
			url:     "https://vimeo.com/watch?v=12345",
			wantErr: true,
			errMsg:  "not a YouTube URL",
		},
		{
			name:    "YouTube URL without watch path",
			url:     "https://www.youtube.com/channel/UC-lHJZR3Gqxm24_Vd_AJ5Yw",
			wantErr: true,
			errMsg:  "not a YouTube watch URL",
		},
		{
			name:    "YouTube URL without video ID",
			url:     "https://www.youtube.com/watch",
			wantErr: true,
			errMsg:  "missing video ID",
		},
		{
			name:    "YouTube URL with invalid video ID length",
			url:     "https://www.youtube.com/watch?v=tooShort",
			wantErr: true,
			errMsg:  "invalid video ID length",
		},
		{
			name:    "YouTube URL with too long video ID",
			url:     "https://www.youtube.com/watch?v=tooLongVideoID123",
			wantErr: true,
			errMsg:  "invalid video ID length",
		},
		{
			name:    "YouTube URL with other query parameters",
			url:     "https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=10s",
			wantErr: false,
		},
	}

	validator := NewYouTubeURLValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.url)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error message = %v, want to contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}
