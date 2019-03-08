package che_devfile

// LANGUAGE_PLUGINS language specific list of plugins needs to configured in che
var LANGUAGE_PLUGINS = map[string][]Tools{
	"go": []Tools{
		Tools{Name: "default", Type: "dockerimage", Image: "eclipse/ubuntu_go", MemoryLimit: "2147483648"},
		Tools{Name: "golang", Type: "chePlugin", ID: "ms-vscode.go:0.9.2"},
	},
	"js": []Tools{
		Tools{Name: "default", Type: "dockerimage", Image: "eclipse/node", MemoryLimit: "2147483648"},
	},
	"nodejs": []Tools{
		Tools{Name: "default", Type: "dockerimage", Image: "eclipse/node", MemoryLimit: "2147483648"},
	},
	"java": []Tools{
		Tools{Name: "default", Type: "dockerimage", Image: "eclipse/ubuntu_jdk8", MemoryLimit: "2147483648"},
		Tools{Name: "java-plugin", Type: "chePlugin", ID: "org.eclipse.che.vscode-redhat.java:0.38.0"},
	},
}
