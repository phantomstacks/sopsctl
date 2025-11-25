package edit

type editCmdOptions struct {
	File               string
	Cluster            string
	DecodeAsEnv        bool
	ShouldDecodeAsFile bool
	DecodeAsFileKey    string
}

func newEditCmdOptions(file string, cluster string, decodeAsEnv bool, decodeAsFile bool, decodeAsFileKey string) *editCmdOptions {
	return &editCmdOptions{
		File:               file,
		Cluster:            cluster,
		DecodeAsEnv:        decodeAsEnv,
		ShouldDecodeAsFile: decodeAsFile,
		DecodeAsFileKey:    decodeAsFileKey,
	}
}
