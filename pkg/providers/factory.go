package providers

import (
	"github.com/raynaythegreat/heron/pkg/auth"
)

var getCredential = auth.GetCredential
var setCredential = auth.SetCredential
var anthropicOAuthConfig = auth.AnthropicOAuthConfig
var refreshAccessToken = auth.RefreshAccessToken
