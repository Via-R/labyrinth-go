package core

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
)

// Check that the file_path is valid
func checkFilePath(file_path string, is_folder bool) error {
	file_path_to_check := file_path
	if is_folder {
		if file_path_array := strings.Split(file_path, "/"); len(file_path_array) > 1 {
			file_path_to_check = strings.Join(file_path_array[:len(file_path_array)-1], "/")
		}
	}

	if _, err := os.Stat(file_path_to_check); err != nil {
		return err
	}

	return nil
}

// Save serialized labyrinth data to file in existing directory
func (f *Field) SaveLabyrinthToFile(file_path string) error {
	if err := checkFilePath(file_path, true); err != nil {
		return f.Error(err.Error())
	}

	serialized_data, err := json.Marshal(f.GetLabyrinth())
	if err != nil {
		return f.Error(err.Error())
	}

	if err := ioutil.WriteFile(file_path, serialized_data, 0644); err != nil {
		return f.Error(err.Error())
	}

	return nil
}

// Deserialize and load labyrinth data from file
func (f *Field) LoadLabyrinthFromFile(file_path string) error {
	if err := checkFilePath(file_path, false); err != nil {
		return f.Error(err.Error())
	}

	serialized_data, err := ioutil.ReadFile(file_path)
	if err != nil {
		return f.Error(err.Error())
	}

	var deserialized_data [][]uint
	if err := json.Unmarshal(serialized_data, &deserialized_data); err != nil {
		return f.Error(err.Error())
	}

	if err := f.LoadLabyrinth(deserialized_data); err != nil {
		return err
	}

	return nil
}
