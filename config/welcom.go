package config

var WelcomeMsgTemplate = `
<!doctype html>
	<html>

	<head>
		<title>DataWay Endpoint</title>
	</head>
	<body>
		<div style="width: 1000px;overflow: hidden;margin: 0 auto;">
	<div style="float:left; margin-top: 50px;width:50%;">

		<pre >
.___  __ /\                            __   .__             ._.
|   _/  |)/______ __  _  _____________|  | _|__| ____   ____| |
|   \   __/  ___/ \ \/ \/ /  _ \_  __ |  |/ |  |/    \ / ___| |
|   ||  | \___ \   \     (  <_> |  | \|    <|  |   |  / /_/  \|
|___||__|/____  >   \/\_/ \____/|__|  |__|_ |__|___|  \___  /__
              \/                           \/       \/_____/ \/

                                    Version: {{.Version}}
                                    OS/Arch: {{.OS}}/{{.Arch}}
                                  ReleaseAt: {{.BuildAt}}
                                     Uptime: {{.Uptime}}
		</pre>
		<p>Welcome to use DataKit.</p>
	</div>
		</div>
	</body>
	</html>
`
