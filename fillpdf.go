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
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// Form represents the PDF form.
// This is a key value map.
type Form map[string]interface{}

// FillFile fills a PDF form with the specified form values and creates a final filled PDF file.
func FillFromReader(form Form, pdfFile io.Reader) (result io.Reader, err error) {
	// Check if the pdftk utility exists.
	_, err = exec.LookPath("pdftk")
	if err != nil {
		return nil, fmt.Errorf("pdftk utility is not installed!")
	}
	fdfFile := createFdfFile(form)
	f, err := os.CreateTemp("", "fdf")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	_, err = f.Write(fdfFile)
	if err != nil {
		return nil, err
	}
	args := []string{
		"-",
		"fill_form", f.Name(),
		"output", "-",
	}
	cmd := exec.Command("pdftk", args...)
	cmd.Stdin = pdfFile
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("pdftk error: %v\nOutput: %s", err, string(out))
	}

	return bytes.NewReader(out), nil
}

// Fill fills a PDF form with the specified form values and creates a final filled PDF file.
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
		return nil, fmt.Errorf("form PDF file does not exist: '%s'", formPDFFile)
	}

	// Check if the pdftk utility exists.
	_, err = exec.LookPath("pdftk")
	if err != nil {
		return nil, fmt.Errorf("pdftk utility is not installed!")
	}

	fdfFile := createFdfFile(form)

	// Create the pdftk command line arguments.
	args := []string{
		formPDFFile,
		"fill_form", "-",
		"output", "-",
	}
	cmd := exec.Command("pdftk", args...)
	cmd.Stdin = bytes.NewReader(fdfFile)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("pdftk error: %v", err)
	}

	return bytes.NewReader(out), nil
}

func createFdfFile(form Form) []byte {
	w := bytes.NewBuffer(nil)

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

	return w.Bytes()
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
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
