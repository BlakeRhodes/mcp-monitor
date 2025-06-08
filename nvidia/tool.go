package nvidia

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

// NewTool creates a NVIDIA GPU information tool
func NewTool() mcp.Tool {
	return mcp.NewTool("get_nvidia_gpu_info",
		mcp.WithDescription("Get NVIDIA GPU hardware and usage information"),
	)
}

// Handler handles NVIDIA GPU information requests
func Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Initialize NVML
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to initialize NVML: %v", nvml.ErrorString(ret))), nil
	}
	defer nvml.Shutdown()

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get device count: %v", nvml.ErrorString(ret))), nil
	}

	gpus := []map[string]interface{}{}

	for i := 0; i < count; i++ {
		dev, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get GPU device at index %d: %v", i, nvml.ErrorString(ret))), nil
		}

		name, _ := dev.GetName()
		uuid, _ := dev.GetUUID()
		serial, _ := dev.GetSerial()
		pci, _ := dev.GetPciInfo()
		mem, _ := dev.GetMemoryInfo()
		util, _ := dev.GetUtilizationRates()
		temp, _ := dev.GetTemperature(nvml.TEMPERATURE_GPU)
		power, _ := dev.GetPowerUsage()
		maxPower, _ := dev.GetEnforcedPowerLimit()

		gpus = append(gpus, map[string]interface{}{
			"name":         name,
			"uuid":         uuid,
			"serial":       serial,
			"pci_bus_id":   pci.BusId,
			"memory_total": mem.Total,
			"memory_used":  mem.Used,
			"memory_free":  mem.Free,
			"gpu_util":     util.Gpu,
			"mem_util":     util.Memory,
			"temperature":  temp,
			"power_watts":  float64(power) / 1000.0,
			"power_limit_watts": float64(maxPower) / 1000.0,
		})
	}

	result := map[string]interface{}{
		"gpus": gpus,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}
