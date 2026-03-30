// Heron - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 Heron contributors

package config

// Runtime environment variable keys for the heron process.
// These control the location of files and binaries at runtime and are read
// directly via os.Getenv / os.LookupEnv. All heron-specific keys use the
// HERON_ prefix. Reference these constants instead of inline string
// literals to keep all supported knobs visible in one place and to prevent
// typos.
const (
	// EnvHome overrides the base directory for all heron data
	// (config, workspace, skills, auth store, …).
	// Default: ~/.heron
	EnvHome = "HERON_HOME"

	// EnvConfig overrides the full path to the JSON config file.
	// Default: $HERON_HOME/config.json
	EnvConfig = "HERON_CONFIG"

	// EnvBuiltinSkills overrides the directory from which built-in
	// skills are loaded.
	// Default: <cwd>/skills
	EnvBuiltinSkills = "HERON_BUILTIN_SKILLS"

	// EnvBinary overrides the path to the heron executable.
	// Used by the web launcher when spawning the gateway subprocess.
	// Default: resolved from the same directory as the current executable.
	EnvBinary = "HERON_BINARY"

	// EnvGatewayHost overrides the host address for the gateway server.
	// Default: "127.0.0.1"
	EnvGatewayHost = "HERON_GATEWAY_HOST"
)
