package disk

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/shirou/gopsutil/disk"
)

type PSDiskStats interface {
	Usage(path string) (*disk.UsageStat, error)
	FilterUsage(mountPointFilter []string, fstypeExclude []string) ([]*disk.UsageStat, []*disk.PartitionStat, error)
	OSGetenv(key string) string
	Partitions(all bool) ([]disk.PartitionStat, error)
}

type PSDisk struct{}

func (dk *PSDisk) Usage(path string) (*disk.UsageStat, error) {
	return disk.Usage(path)
}

func (dk *PSDisk) OSGetenv(key string) string {
	return os.Getenv(key)
}

func (dk *PSDisk) Partitions(all bool) ([]disk.PartitionStat, error) {
	return disk.Partitions(all)
}

func (dk *PSDisk) FilterUsage(mountPointFilter []string, fstypeExclude []string,
) ([]*disk.UsageStat, []*disk.PartitionStat, error) {
	parts, err := dk.Partitions(true)
	if err != nil {
		return nil, nil, err
	}

	// Make a "set" out of the filter slice
	mountPointFilterSet := make(map[string]bool)
	for _, filter := range mountPointFilter {
		mountPointFilterSet[filter] = true
	}
	fstypeExcludeSet := make(map[string]bool)
	for _, filter := range fstypeExclude {
		fstypeExcludeSet[filter] = true
	}
	paths := make(map[string]bool)
	for _, part := range parts {
		paths[part.Mountpoint] = true
	}

	// Autofs mounts indicate a potential mount, the partition will also be
	// listed with the actual filesystem when mounted.  Ignore the autofs
	// partition to avoid triggering a mount.
	fstypeExcludeSet["autofs"] = true

	var usage []*disk.UsageStat
	var partitions []*disk.PartitionStat
	hostMountPrefix := dk.OSGetenv("HOST_MOUNT_PREFIX")

	for i := range parts {
		p := parts[i]

		if len(mountPointFilter) > 0 {
			// If the mount point is not a member of the filter set,
			// don't gather info on it.
			if _, ok := mountPointFilterSet[p.Mountpoint]; !ok {
				continue
			}
		}

		// If the mount point is a member of the exclude set,
		// don't gather info on it.
		if _, ok := fstypeExcludeSet[p.Fstype]; ok {
			continue
		}

		// If there's a host mount prefix, exclude any paths which conflict
		// with the prefix.
		if len(hostMountPrefix) > 0 &&
			!strings.HasPrefix(p.Mountpoint, hostMountPrefix) &&
			paths[hostMountPrefix+p.Mountpoint] {
			continue
		}

		du, err := dk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}

		du.Path = filepath.Join("/", strings.TrimPrefix(p.Mountpoint, hostMountPrefix))
		du.Fstype = p.Fstype
		usage = append(usage, du)
		partitions = append(partitions, &p)
	}

	return usage, partitions, nil
}

type MountOptions []string

func (opts MountOptions) Mode() string {
	switch {
	case opts.exists("rw"):
		return "rw"
	case opts.exists("ro"):
		return "ro"
	default:
		return "unknown"
	}
}

func (opts MountOptions) exists(opt string) bool {
	for _, o := range opts {
		if o == opt {
			return true
		}
	}
	return false
}

func parseOptions(opts string) MountOptions {
	return strings.Split(opts, ",")
}
