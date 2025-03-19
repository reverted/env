package env

import (
	"os"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		input       any
		expected    any
		expectError bool
	}{
		{
			name: "basic string parsing",
			envVars: map[string]string{
				"APP_NAME": "test-app",
			},
			input: &struct {
				AppName string `env:"APP_NAME"`
			}{},
			expected: &struct {
				AppName string `env:"APP_NAME"`
			}{
				AppName: "test-app",
			},
		},
		{
			name: "basic int parsing",
			envVars: map[string]string{
				"PORT": "8080",
			},
			input: &struct {
				Port int `env:"PORT"`
			}{},
			expected: &struct {
				Port int `env:"PORT"`
			}{
				Port: 8080,
			},
		},
		{
			name: "string slice parsing",
			envVars: map[string]string{
				"ALLOWED_ORIGINS": "http://localhost:3000,http://example.com",
			},
			input: &struct {
				AllowedOrigins []string `env:"ALLOWED_ORIGINS"`
			}{},
			expected: &struct {
				AllowedOrigins []string `env:"ALLOWED_ORIGINS"`
			}{
				AllowedOrigins: []string{"http://localhost:3000", "http://example.com"},
			},
		},
		{
			name: "int slice parsing",
			envVars: map[string]string{
				"ALLOWED_PORTS": "8080,9000,3000",
			},
			input: &struct {
				AllowedPorts []int `env:"ALLOWED_PORTS"`
			}{},
			expected: &struct {
				AllowedPorts []int `env:"ALLOWED_PORTS"`
			}{
				AllowedPorts: []int{8080, 9000, 3000},
			},
		},
		{
			name:    "missing required env",
			envVars: map[string]string{},
			input: &struct {
				Required string `env:"REQUIRED_VAR"`
			}{},
			expectError: true,
		},
		{
			name:    "optional env",
			envVars: map[string]string{},
			input: &struct {
				Optional string `env:"OPTIONAL_VAR,optional"`
			}{},
			expected: &struct {
				Optional string `env:"OPTIONAL_VAR,optional"`
			}{
				Optional: "",
			},
		},
		{
			name:    "default value",
			envVars: map[string]string{},
			input: &struct {
				WithDefault string `env:"WITH_DEFAULT,default=default-value"`
			}{},
			expected: &struct {
				WithDefault string `env:"WITH_DEFAULT,default=default-value"`
			}{
				WithDefault: "default-value",
			},
		},
		{
			name: "env overrides default",
			envVars: map[string]string{
				"WITH_DEFAULT": "env-value",
			},
			input: &struct {
				WithDefault string `env:"WITH_DEFAULT,default=default-value"`
			}{},
			expected: &struct {
				WithDefault string `env:"WITH_DEFAULT,default=default-value"`
			}{
				WithDefault: "env-value",
			},
		},
		{
			name: "invalid int format",
			envVars: map[string]string{
				"PORT": "invalid-port",
			},
			input: &struct {
				Port int `env:"PORT"`
			}{},
			expectError: true,
		},
		{
			name: "invalid int slice element",
			envVars: map[string]string{
				"PORTS": "8080,invalid,9000",
			},
			input: &struct {
				Ports []int `env:"PORTS"`
			}{},
			expectError: true,
		},
		{
			name:    "invalid tag format",
			envVars: map[string]string{},
			input: &struct {
				Invalid string `env:"INVALID,default"`
			}{},
			expectError: true,
		},
		{
			name: "combination of types",
			envVars: map[string]string{
				"APP_NAME":      "test-app",
				"PORT":          "8080",
				"DEBUG":         "true",
				"ALLOWED_HOSTS": "localhost,example.com",
			},
			input: &struct {
				AppName      string   `env:"APP_NAME"`
				Port         int      `env:"PORT"`
				Debug        string   `env:"DEBUG"`
				AllowedHosts []string `env:"ALLOWED_HOSTS"`
				Optional     string   `env:"OPTIONAL,optional"`
				DefaultVal   string   `env:"DEFAULT_VAL,default=default"`
			}{},
			expected: &struct {
				AppName      string   `env:"APP_NAME"`
				Port         int      `env:"PORT"`
				Debug        string   `env:"DEBUG"`
				AllowedHosts []string `env:"ALLOWED_HOSTS"`
				Optional     string   `env:"OPTIONAL,optional"`
				DefaultVal   string   `env:"DEFAULT_VAL,default=default"`
			}{
				AppName:      "test-app",
				Port:         8080,
				Debug:        "true",
				AllowedHosts: []string{"localhost", "example.com"},
				Optional:     "",
				DefaultVal:   "default",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// Run Parse
			err := Parse(tt.input)

			// Check error
			if (err != nil) != tt.expectError {
				t.Errorf("Parse() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError && !reflect.DeepEqual(tt.input, tt.expected) {
				t.Errorf("Parse() got = %v, want %v", tt.input, tt.expected)
			}
		})
	}
}

func TestParseTag(t *testing.T) {
	tests := []struct {
		name        string
		tag         reflect.StructTag
		expected    Tag
		expectedOk  bool
		expectError bool
	}{
		{
			name:       "basic tag",
			tag:        `env:"APP_NAME"`,
			expected:   Tag{Env: "APP_NAME", Optional: false, Default: ""},
			expectedOk: true,
		},
		{
			name:       "optional tag",
			tag:        `env:"APP_NAME,optional"`,
			expected:   Tag{Env: "APP_NAME", Optional: true, Default: ""},
			expectedOk: true,
		},
		{
			name:       "default tag",
			tag:        `env:"APP_NAME,default=test"`,
			expected:   Tag{Env: "APP_NAME", Optional: false, Default: "test"},
			expectedOk: true,
		},
		{
			name:       "combined options",
			tag:        `env:"APP_NAME,optional,default=test"`,
			expected:   Tag{Env: "APP_NAME", Optional: true, Default: "test"},
			expectedOk: true,
		},
		{
			name:       "no tag",
			tag:        `json:"name"`,
			expected:   Tag{},
			expectedOk: false,
		},
		{
			name:        "empty tag",
			tag:         `env:""`,
			expectError: true,
		},
		{
			name:        "invalid default format",
			tag:         `env:"APP_NAME,default"`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag, ok, err := parseTag(tt.tag)

			if (err != nil) != tt.expectError {
				t.Errorf("parseTag() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if ok != tt.expectedOk {
				t.Errorf("parseTag() ok = %v, expectedOk %v", ok, tt.expectedOk)
				return
			}

			if !tt.expectError && ok && !reflect.DeepEqual(tag, tt.expected) {
				t.Errorf("parseTag() tag = %v, expected %v", tag, tt.expected)
			}
		})
	}
}

func TestParseSlice(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		elemType    reflect.Type
		expected    interface{}
		expectError bool
	}{
		{
			name:     "string slice",
			value:    "a,b,c",
			elemType: reflect.TypeOf(""),
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "int slice",
			value:    "1,2,3",
			elemType: reflect.TypeOf(0),
			expected: []int{1, 2, 3},
		},
		{
			name:        "invalid int element",
			value:       "1,x,3",
			elemType:    reflect.TypeOf(0),
			expectError: true,
		},
		{
			name:        "unsupported type",
			value:       "1.1,2.2",
			elemType:    reflect.TypeOf(1.1),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSlice(tt.value, tt.elemType)

			if (err != nil) != tt.expectError {
				t.Errorf("parseSlice() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError && !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseSlice() result = %v, expected %v", result, tt.expected)
			}
		})
	}
}
