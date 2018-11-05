/*
 *  FillPDF - Fill PDF forms
 *  Copyright DesertBit
 *  Author: Roland Singer
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package fillpdf

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// Form represents the PDF form.
// This is a key value map.
type Form map[string]interface{}

// Fill a PDF form with the specified form values and create a final filled PDF file.
func Fill(form Form, formPDFFile string) (result io.Reader, err error) {
	// Get the absolute paths.
	formPDFFile, err = filepath.Abs(formPDFFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create the absolute path: %v", err)
	}

	// Check if the form file exists.
	e, err := exists(formPDFFile)
	if err != nil {
		return nil, fmt.Errorf("failed to check if form PDF file exists: %v", err)
	} else if !e {
		return nil, fmt.Errorf("form PDF file does not exists: '%s'", formPDFFile)
	}

	// Check if the pdftk utility exists.
	_, err = exec.LookPath("pdftk")
	if err != nil {
		return nil, fmt.Errorf("pdftk utility is not installed!")
	}

	// Create a temporary directory.
	tmpDir, err := ioutil.TempDir("", "fillpdf-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %v", err)
	}

	// Remove the temporary directory on defer again.
	defer func() {
		errD := os.RemoveAll(tmpDir)
		// Log the error only.
		if errD != nil {
			log.Printf("fillpdf: failed to remove temporary directory '%s' again: %v", tmpDir, errD)
		}
	}()

	// Create the fdf data file.
	fdfFile := filepath.Clean(tmpDir + "/data.fdf")
	err = createFdfFile(form, fdfFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create fdf form data file: %v", err)
	}

	// Create the pdftk command line arguments.
	args := []string{
		formPDFFile,
		"fill_form", fdfFile,
		"output", "-",
	}

	// Run the pdftk utility.
	out, err := runCommandInPath(tmpDir, "pdftk", args...)
	if err != nil {
		return nil, fmt.Errorf("pdftk error: %v", err)
	}

	return out, nil
}

func createFdfFile(form Form, path string) error {
	// Create the file.
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a new writer.
	w := bufio.NewWriter(file)

	// Write the fdf header.
	fmt.Fprintln(w, fdfHeader)

	// Write the form data.
	for key, value := range form {
		var valStr string
		switch v := value.(type) {
		case bool:
			if v {
				valStr = "Yes"
			} else {
				valStr = "Off"
			}
		case float64:
			valStr = strconv.FormatFloat(v, 'f', -1, 64)
		default:
			valStr = fmt.Sprintf("%v", value)
		}
		fmt.Fprintf(w, "<< /T (%s) /V (%s)>>\n", key, valStr)
	}

	// Write the fdf footer.
	fmt.Fprintln(w, fdfFooter)

	// Flush everything.
	return w.Flush()
}

const fdfHeader = `%FDF-1.2
%,,oe"
1 0 obj
<<
/FDF << /Fields [`

const fdfFooter = `]
>>
>>
endobj
trailer
<<
/Root 1 0 R
>>
%%EOF`
