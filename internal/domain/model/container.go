package model

import (
	"time"

	"github.com/containers/podman/v5/pkg/domain/entities"
)

type ContainerEntity struct {
	Name      string                 `json:"name"`
	Path      string                 `json:"path"`
	Status    string                 `json:"status"` // active, inactive, failed, etc.
	Container entities.ListContainer `json:"container,omitzero"`
}

type ContainerStats struct {
	Read         time.Time          `json:"read"`
	Preread      time.Time          `json:"preread"`
	PidsStats    PidsStats          `json:"pids_stats"`
	CPUStats     CPUStats           `json:"cpu_stats"`
	PreCPU       CPUStats           `json:"precpu_stats"`
	Memory       MemoryStats        `json:"memory_stats"`
	Networks     map[string]Network `json:"networks"`
	Name         string             `json:"name"`
	ID           string             `json:"Id"`
	NumProcs     int                `json:"num_procs"`
	BlkioStats   BlkioStats         `json:"blkio_stats"`
	StorageStats StorageStats       `json:"storage_stats"`
}

type PidsStats struct {
	Current uint64 `json:"current"`
}

type CPUStats struct {
	CPUUsage       CPUUsage       `json:"cpu_usage"`
	SystemCPUUsage uint64         `json:"system_cpu_usage"`
	OnlineCPUs     int            `json:"online_cpus"`
	ThrottlingData ThrottlingData `json:"throttling_data"`
}

type CPUUsage struct {
	TotalUsage        uint64 `json:"total_usage"`
	UsageInKernelmode uint64 `json:"usage_in_kernelmode"`
	UsageInUsermode   uint64 `json:"usage_in_usermode"`
}

type ThrottlingData struct {
	Periods          uint64 `json:"periods"`
	ThrottledPeriods uint64 `json:"throttled_periods"`
	ThrottledTime    uint64 `json:"throttled_time"`
}

type MemoryStats struct {
	Usage uint64 `json:"usage"`
	Limit uint64 `json:"limit"`
}

type Network struct {
	RxBytes   uint64 `json:"rx_bytes"`
	RxPackets uint64 `json:"rx_packets"`
	RxErrors  uint64 `json:"rx_errors"`
	RxDropped uint64 `json:"rx_dropped"`
	TxBytes   uint64 `json:"tx_bytes"`
	TxPackets uint64 `json:"tx_packets"`
	TxErrors  uint64 `json:"tx_errors"`
	TxDropped uint64 `json:"tx_dropped"`
}

type BlkioStats struct {
	IoServiceBytesRecursive []BlkioStatEntry `json:"io_service_bytes_recursive"`
	IoServicedRecursive     []BlkioStatEntry `json:"io_serviced_recursive"`
	IoQueueRecursive        []BlkioStatEntry `json:"io_queue_recursive"`
	IoServiceTimeRecursive  []BlkioStatEntry `json:"io_service_time_recursive"`
	IoWaitTimeRecursive     []BlkioStatEntry `json:"io_wait_time_recursive"`
	IoMergedRecursive       []BlkioStatEntry `json:"io_merged_recursive"`
	IoTimeRecursive         []BlkioStatEntry `json:"io_time_recursive"`
	SectorsRecursive        []BlkioStatEntry `json:"sectors_recursive"`
}

type BlkioStatEntry struct {
	Major uint64 `json:"major"`
	Minor uint64 `json:"minor"`
	Op    string `json:"op"` // "Read", "Write", "Sync", "Async", "Total"
	Value uint64 `json:"value"`
}

type StorageStats struct {
	ReadCount  uint64 `json:"read_count,omitempty"`
	ReadSize   uint64 `json:"read_size,omitempty"`
	WriteCount uint64 `json:"write_count,omitempty"`
	WriteSize  uint64 `json:"write_size,omitempty"`
}
