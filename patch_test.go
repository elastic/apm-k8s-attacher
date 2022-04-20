package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestJSONPatches(t *testing.T) {
	var patches []patchOperation

	patches = append(patches, createVolumePatch(false))
	patches = append(patches, createInitContainerPatch(true))
	patches = append(patches, createVolumeMountsPatch(true, 1))
	patches = append(patches, createEnvVariablesPatches(true, 2)...)

	patchBytes, err := json.MarshalIndent(patches, "", "    ")
	if err != nil {
		fmt.Printf("marshalling error: %s", err)
	}
	json := string(patchBytes)
	fmt.Printf("volume:\n%+v", json)
}
