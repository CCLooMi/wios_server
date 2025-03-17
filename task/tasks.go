package task

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Invoke(startDZIServer),
	fx.Invoke(startVideoProcessor),
	fx.Invoke(startUploadCleaner),
)
