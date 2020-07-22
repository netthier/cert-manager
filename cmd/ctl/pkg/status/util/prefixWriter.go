/*
Copyright 2020 The Jetstack cert-manager contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/kubectl/pkg/describe"
	"k8s.io/kubectl/pkg/util/event"
)

// This file defines implementation of the PrefixWriter interface defined in "k8s.io/kubectl/pkg/describe"
// This implementation is based on the one in the describe package, with a slight modification of having a baseLevel
// on top of which any other indentations are added.
// The purpose is be able to reuse functions in the describe package where the Level of the output is fixed,
// e.g. DescribeEvents() only prints out at Level 0.

// prefixWriter implements describe.PrefixWriter
type prefixWriter struct {
	out       *tabwriter.Writer
	baseLevel int
}

var _ describe.PrefixWriter = &prefixWriter{}

// NewPrefixWriter creates a new PrefixWriter.
func NewPrefixWriter(out *tabwriter.Writer) *prefixWriter {
	return &prefixWriter{out: out, baseLevel: 0}
}

func (pw *prefixWriter) Write(level int, format string, a ...interface{}) {
	level += pw.baseLevel
	levelSpace := "  "
	prefix := ""
	for i := 0; i < level; i++ {
		prefix += levelSpace
	}
	fmt.Fprintf(pw.out, prefix+format, a...)
}

func (pw *prefixWriter) WriteLine(a ...interface{}) {
	fmt.Fprintln(pw.out, a...)
}

func (pw *prefixWriter) Flush() {
	pw.out.Flush()
}

func DescribeEvents(el *corev1.EventList, w describe.PrefixWriter, baseLevel int) {
	if len(el.Items) == 0 {
		w.Write(baseLevel, "Events:\t<none>\n")
		w.Flush()
		return
	}
	w.Flush()
	sort.Sort(event.SortableEvents(el.Items))
	w.Write(baseLevel, "Events:\n")
	w.Write(baseLevel+1, "Type\tReason\tAge\tFrom\tMessage\n")
	w.Write(baseLevel+1, "----\t------\t----\t----\t-------\n")
	for _, e := range el.Items {
		var interval string
		if e.Count > 1 {
			interval = fmt.Sprintf("%s (x%d over %s)", translateTimestampSince(e.LastTimestamp), e.Count, translateTimestampSince(e.FirstTimestamp))
		} else {
			interval = translateTimestampSince(e.FirstTimestamp)
		}
		w.Write(baseLevel+1, "%v\t%v\t%s\t%v\t%v\n",
			e.Type,
			e.Reason,
			interval,
			formatEventSource(e.Source),
			strings.TrimSpace(e.Message),
		)
	}
	w.Flush()
}

// formatEventSource formats EventSource as a comma separated string excluding Host when empty
func formatEventSource(es corev1.EventSource) string {
	EventSourceString := []string{es.Component}
	if len(es.Host) > 0 {
		EventSourceString = append(EventSourceString, es.Host)
	}
	return strings.Join(EventSourceString, ", ")
}

// translateTimestampSince returns the elapsed time since timestamp in
// human-readable approximation.
func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}
