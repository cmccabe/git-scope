/**
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

func main() {
	app := kingpin.New("git-scope", "git branch comparison tool")
	app.HelpFlag.Short('h')
	diff := app.Command("diff", "Show the differences between two branches.")
	srcBranch := diff.Arg("src", "Source branch").Required().String()
	dstBranch := diff.Arg("dest", "Destination branch").Required().String()

	switch (kingpin.MustParse(app.Parse(os.Args[1:]))) {
	case "diff":
		err := doDiff(os.Stdout, *srcBranch, *dstBranch)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
	}
}
