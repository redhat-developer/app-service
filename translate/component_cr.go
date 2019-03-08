package translate

type ComponentCr struct {
	ApiVersion string
	Kind       string
	Metadata   struct {
		Name string
	}
	Spec struct {
		App        string
		Buildtype  string
		Codebase   string
		Listenport string
	}
}
