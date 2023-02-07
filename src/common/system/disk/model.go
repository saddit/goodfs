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

package disk

import "common/datasize"

//go:generate msgp -tests=false

// Info stat fs struct is container which holds following values
// Total - total size of the volume / disk
// Free - free size of the volume / disk
// Files - total inodes available
// Ffree - free inodes available
// FSType - file system type
type Info struct {
	Total  datasize.DataSize `msg:"total" json:"total"`
	Free   datasize.DataSize `msg:"free" json:"free"`
	Used   datasize.DataSize `msg:"used" json:"used"`
	Files  uint64            `msg:"files" json:"files"`
	Ffree  uint64            `msg:"f_free" json:"ffree"`
	Major  uint32            `msg:"major" json:"major"`
	Minor  uint32            `msg:"minor" json:"minor"`
	FSType string            `msg:"fs_type" json:"fsType"`
}

// DevID is the device name
type DevID string

// AllDrivesIOStats is map between drive devices and IO stats
type AllDrivesIOStats map[DevID]*IOStats

// IOStats contains stats of a single drive
type IOStats struct {
	ReadBytes  datasize.DataSize `json:"readBytes" msg:"read_bytes"`
	WriteBytes datasize.DataSize `json:"writeBytes" msg:"write_bytes"`
	ReadCount  uint64            `json:"readCount" msg:"read_count"`
	WriteCount uint64            `json:"writeCount" msg:"write_count"`
	ReadTime   uint64            `json:"readTime" msg:"read_time"`
	WriteTime  uint64            `json:"writeTime" msg:"write_time"`
	CurrentIOs uint64            `json:"currentIOs" msg:"current_ios"` // CurrentIOs only valid in linux.
	IoTime     uint64            `json:"ioTime" msg:"io_time"`         // IoTime only valid in linux.
	WeightedIO uint64            `json:"weightedIO" msg:"weighted_io"` // WeightedIO only valid in linux.
}
