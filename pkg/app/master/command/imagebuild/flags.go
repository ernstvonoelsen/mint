package imagebuild

import (
	"runtime"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// ImageBuild command flag names and usage descriptions
const (
	FlagRegistryPush      = "registry-push"
	FlagRegistryPushUsage = "Push the built image to a container registry"

	FlagRuntimeLoad      = "runtime-load"
	FlagRuntimeLoadUsage = "container runtime where to load to created image"

	FlagEngine      = "engine"
	FlagEngineUsage = "container image build engine to use"

	FlagEngineEndpoint      = "engine-endpoint"
	FlagEngineEndpointUsage = "build engine endpoint address"

	FlagEngineToken      = "engine-token"
	FlagEngineTokenUsage = "build engine specific API token"

	FlagEngineNamespace      = "engine-namespace"
	FlagEngineNamespaceUsage = "build engine specific namespace"

	FlagImageName      = "image-name"
	FlagImageNameUsage = "container image name to use (including tag)"

	FlagImageArchiveFile      = "image-archive-file"
	FlagImageArchiveFileUsage = "local file path for the image tar archive file"

	FlagDockerfile      = "dockerfile"
	FlagDockerfileUsage = "local Dockerfile path (for buildkit and depot) or a relative to the build context directory (for docker or podman)"

	FlagContextDir      = "context-dir"
	FlagContextDirUsage = "local build context directory"

	FlagBuildArg      = "build-arg"
	FlagBuildArgUsage = "build time variable (ARG)"

	FlagLabel      = "label"
	FlagLabelUsage = "image label to add"

	FlagArchitecture      = "architecture"
	FlagArchitectureUsage = "build architecture"

	FlagBase      = "base"
	FlagBaseUsage = "base image to use (from selected runtime, docker by default, or pulled if not available)"

	FlagBaseTar      = "base-tar"
	FlagBaseTarUsage = "base image from a local tar file"

	FlagBaseWithCerts      = "base-with-certs"
	FlagBaseWithCertsUsage = "static-debian12 distroless base image - contains only certs and timezone info"

	FlagExePath      = "exe-path"
	FlagExePathUsage = "local (linux) executable file that will be used as the entrypoint for the new image (added to the selected base image or scratch image if no base image is provided)"
)

const (
	DefaultImageName        = "mint-new-container-image:latest"
	DefaultImageArchiveFile = "mint-new-container-image.tar"
	DefaultDockerfilePath   = "Dockerfile"
	DefaultContextDir       = "."

	NoneRuntimeLoad   = "none"
	DockerRuntimeLoad = "docker"
	PodmanRuntimeLoad = "podman"

	DockerBuildEngine       = "docker"
	DockerBuildEngineInfo   = "Native Docker container build engine"
	BuildkitBuildEngine     = "buildkit"
	BuildkitBuildEngineInfo = "BuildKit container build engine"
	DepotBuildEngine        = "depot"
	DepotBuildEngineInfo    = "Depot.dev cloud-based container build engine"
	PodmanBuildEngine       = "podman"
	PodmanBuildEngineInfo   = "Native Podman/Buildah container build engine"
	SimpleBuildEngine       = "simple"
	SimpleBuildEngineInfo   = "Built-in container build engine for simple images that do not use 'RUN' instructions"

	Amd64Arch = "amd64"
	Arm64Arch = "arm64"

	DefaultRuntimeLoad = NoneRuntimeLoad
	DefaultEngineName  = DockerBuildEngine
)

func GetDefaultBuildArch() string {
	switch runtime.GOARCH {
	case Amd64Arch:
		return Amd64Arch
	case Arm64Arch:
		return Arm64Arch
	default:
		return Amd64Arch
	}
}

type BuildEngineProps struct {
	Info                  string
	TokenRequired         bool
	NamespaceRequired     bool
	EndpointRequired      bool
	NativeTokenEnvVar     string
	NativeNamespaceEnvVar string
	NativeNamespaceName   string
	DefaultEndpoint       string
}

var Runtimes = map[string]struct{}{
	NoneRuntimeLoad:   {},
	DockerRuntimeLoad: {},
	PodmanRuntimeLoad: {},
}

func IsRuntimeValue(name string) bool {
	if _, ok := Runtimes[name]; ok {
		return true
	}

	return false
}

var Architectures = map[string]struct{}{
	Amd64Arch: {},
	Arm64Arch: {},
}

func IsArchValue(name string) bool {
	if _, ok := Architectures[name]; ok {
		return true
	}

	return false
}

var BuildEngines = map[string]BuildEngineProps{
	DockerBuildEngine: {Info: DockerBuildEngineInfo},
	BuildkitBuildEngine: {
		Info:             BuildkitBuildEngineInfo,
		EndpointRequired: true,
	},
	DepotBuildEngine: {
		Info:                  DepotBuildEngineInfo,
		TokenRequired:         true,
		NamespaceRequired:     true,
		NativeTokenEnvVar:     "DEPOT_TOKEN",
		NativeNamespaceEnvVar: "DEPOT_PROJECT_ID",
		NativeNamespaceName:   "project",
	},
	PodmanBuildEngine: {Info: PodmanBuildEngineInfo},
	SimpleBuildEngine: {Info: SimpleBuildEngineInfo},
}

var Flags = map[string]cli.Flag{
	FlagEngine: &cli.StringFlag{
		Name:    FlagEngine,
		Value:   DefaultEngineName,
		Usage:   FlagEngineUsage,
		EnvVars: []string{"DSLIM_IMAGEBUILD_ENGINE"},
	},
	FlagEngineEndpoint: &cli.StringFlag{
		Name:    FlagEngineEndpoint,
		Value:   "",
		Usage:   FlagEngineEndpointUsage,
		EnvVars: []string{"DSLIM_IMAGEBUILD_ENGINE_ENDPOINT"},
	},
	FlagEngineToken: &cli.StringFlag{
		Name:    FlagEngineToken,
		Value:   "",
		Usage:   FlagEngineTokenUsage,
		EnvVars: []string{"DSLIM_IMAGEBUILD_ENGINE_TOKEN"},
	},
	FlagEngineNamespace: &cli.StringFlag{
		Name:    FlagEngineNamespace,
		Value:   "",
		Usage:   FlagEngineNamespaceUsage,
		EnvVars: []string{"DSLIM_IMAGEBUILD_ENGINE_NS"},
	},
	FlagImageName: &cli.StringFlag{
		Name:    FlagImageName,
		Value:   DefaultImageName,
		Usage:   FlagImageNameUsage,
		EnvVars: []string{"DSLIM_IMAGEBUILD_IMAGE_NAME"},
	},
	FlagImageArchiveFile: &cli.StringFlag{
		Name:    FlagImageArchiveFile,
		Value:   DefaultImageArchiveFile,
		Usage:   FlagImageArchiveFileUsage,
		EnvVars: []string{"DSLIM_IMAGEBUILD_IMAGE_ARCHIVE"},
	},
	FlagDockerfile: &cli.StringFlag{
		Name:    FlagDockerfile,
		Value:   DefaultDockerfilePath,
		Usage:   FlagDockerfileUsage,
		EnvVars: []string{"DSLIM_IMAGEBUILD_DOCKERFILE"},
	},
	FlagContextDir: &cli.StringFlag{
		Name:    FlagContextDir,
		Value:   DefaultContextDir,
		Usage:   FlagContextDirUsage,
		EnvVars: []string{"DSLIM_IMAGEBUILD_CONTEXT_DIR"},
	},
	FlagBuildArg: &cli.StringSliceFlag{
		Name:    FlagBuildArg,
		Value:   cli.NewStringSlice(""),
		Usage:   FlagBuildArgUsage,
		EnvVars: []string{"DSLIM_IMAGEBUILD_BUILD_ARGS"},
	},
	FlagLabel: &cli.StringSliceFlag{
		Name:    FlagLabel,
		Value:   cli.NewStringSlice(""),
		Usage:   FlagLabelUsage,
		EnvVars: []string{"DSLIM_IMAGEBUILD_LABELS"},
	},
	FlagRuntimeLoad: &cli.StringSliceFlag{
		Name:    FlagRuntimeLoad,
		Usage:   FlagRuntimeLoadUsage,
		EnvVars: []string{"DSLIM_IMAGEBUILD_RUNTIME_LOAD"},
	},
	FlagArchitecture: &cli.StringFlag{
		Name:    FlagArchitecture,
		Value:   GetDefaultBuildArch(),
		Usage:   FlagArchitectureUsage,
		EnvVars: []string{"DSLIM_IMAGEBUILD_ARCH"},
	},
	FlagBase: &cli.StringFlag{
		Name:    FlagBase,
		Value:   "",
		Usage:   FlagBaseUsage,
		EnvVars: []string{"DSLIM_IMAGEBUILD_BASE"},
	},
	FlagBaseTar: &cli.StringFlag{
		Name:    FlagBaseTar,
		Value:   "",
		Usage:   FlagBaseTarUsage,
		EnvVars: []string{"DSLIM_IMAGEBUILD_BASE_TAR"},
	},
	FlagBaseWithCerts: &cli.BoolFlag{
		Name:    FlagBaseWithCerts,
		Usage:   FlagBaseWithCertsUsage,
		EnvVars: []string{"DSLIM_IMAGEBUILD_BASE_WITH_CERTS"},
	},
	FlagExePath: &cli.StringFlag{
		Name:    FlagExePath,
		Value:   "",
		Usage:   FlagExePathUsage,
		EnvVars: []string{"DSLIM_IMAGEBUILD_EXE_PATH"},
	},
	FlagRegistryPush: &cli.BoolFlag{
		Name:    FlagRegistryPush,
		Value:   false,
		Usage:   FlagRegistryPushUsage,
		EnvVars: []string{"DSLIM_IMAGEBUILD_REGISTRY_PUSH"},
	},
}

func cflag(name string) cli.Flag {
	cf, ok := Flags[name]
	if !ok {
		log.Fatalf("unknown flag='%s'", name)
	}

	return cf
}

func useAllFlags() []cli.Flag {
	var flagList []cli.Flag
	for k := range Flags {
		flagList = append(flagList, cflag(k))
	}

	return flagList
}
