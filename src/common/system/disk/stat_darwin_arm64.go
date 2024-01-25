//go:build darwin && arm64

package disk

// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

import (
	"common/datasize"
	"syscall"

	diskutil "github.com/shirou/gopsutil/v3/disk"
	"golang.org/x/sys/unix"
)

var (
	Root = "/"
)

// GetInfo returns total and free bytes available in a directory, e.g. `/`.
func GetInfo(path string) (info Info, err error) {
	stats, err := diskutil.Usage(path)
	if err != nil {
		return Info{}, err
	}
	info = Info{
		Total:  datasize.DataSize(stats.Total),
		Free:   datasize.DataSize(stats.Free),
		Used:   datasize.DataSize(stats.Used),
		Files:  stats.InodesTotal,
		Ffree:  stats.InodesFree,
		FSType: stats.Fstype,
	}

	st := syscall.Stat_t{}
	err = syscall.Stat(path, &st)
	if err != nil {
		return Info{}, err
	}
	//nolint:unconvert
	devID := uint64(st.Dev) // Needed to support multiple GOARCHs
	info.Major = unix.Major(devID)
	info.Minor = unix.Minor(devID)
	return info, nil
}
