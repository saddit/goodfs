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
	FSType string            `msg:"fs_type" json:"fsType"`
	Major  uint32            `msg:"major" json:"major"`
	Minor  uint32            `msg:"minor" json:"minor"`
}

// DevID is the drive major and minor ids
type DevID struct {
	Major uint32
	Minor uint32
}

// AllDrivesIOStats is map between drive devices and IO stats
type AllDrivesIOStats map[DevID]IOStats

// IOStats contains stats of a single drive
type IOStats struct {
	ReadIOs        uint64
	ReadMerges     uint64
	ReadSectors    uint64
	ReadTicks      uint64
	WriteIOs       uint64
	WriteMerges    uint64
	WriteSectors   uint64
	WriteTicks     uint64
	CurrentIOs     uint64
	TotalTicks     uint64
	ReqTicks       uint64
	DiscardIOs     uint64
	DiscardMerges  uint64
	DiscardSectors uint64
	DiscardTicks   uint64
	FlushIOs       uint64
	FlushTicks     uint64
}
