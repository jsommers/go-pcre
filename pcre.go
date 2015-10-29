// Copyright (c) 2011 Florian Weimer. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
// * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright
// notice, this list of conditions and the following disclaimer in the
// documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// This package provides access to the Perl Compatible Regular
// Expresion library, PCRE.
//
// It implements two main types, Regexp and Matcher. Regexp objects
// store a compiled regular expression. They consist of two immutable
// parts: pcre and pcre_extra. You can add pcre_exta to Compiled Regexp by
// studying it with Study() function.
// Compilation of regular expressions using Compile or MustCompile is
// slightly expensive, so these objects should be kept and reused,
// instead of compiling them from scratch for each matching attempt.
// CompileJIT and MustCompileJIT are way more expensive then ordinary
// methods, becose they run Study() func after Regexp compiled but gives
// much better perfomance:
// http://sljit.sourceforge.net/regex_perf.html
//
// Matcher objects keeps the results of a match against a []byte or
// string subject. The Group and GroupString functions provide access
// to capture groups; both versions work no matter if the subject was a
// []byte or string.
//
// Matcher objects contain some temporary space and refer the original
// subject. They are mutable and can be reused (using Match,
// MatchString, Reset or ResetString).
//
// Most of Matcher.*String method are just links to []byte methods, so keep
// this in mind.
//
// For details on the regular expression language implemented by this
// package and the flags defined below, see the PCRE documentation.
// http://www.pcre.org/pcre.txt
package pcre

/*
#cgo pkg-config: libpcre
#include <pcre.h>
#include <string.h>
*/
import "C"

import (
	"fmt"
	"strconv"
	"unsafe"
)

// Flags for Compile and Match functions.
const (
	ANCHORED        = C.PCRE_ANCHORED
	BSR_ANYCRLF     = C.PCRE_BSR_ANYCRLF
	BSR_UNICODE     = C.PCRE_BSR_UNICODE
	NEWLINE_ANY     = C.PCRE_NEWLINE_ANY
	NEWLINE_ANYCRLF = C.PCRE_NEWLINE_ANYCRLF
	NEWLINE_CR      = C.PCRE_NEWLINE_CR
	NEWLINE_CRLF    = C.PCRE_NEWLINE_CRLF
	NEWLINE_LF      = C.PCRE_NEWLINE_LF
	NO_UTF8_CHECK   = C.PCRE_NO_UTF8_CHECK
)

// Flags for Compile functions
const (
	CASELESS          = C.PCRE_CASELESS
	DOLLAR_ENDONLY    = C.PCRE_DOLLAR_ENDONLY
	DOTALL            = C.PCRE_DOTALL
	DUPNAMES          = C.PCRE_DUPNAMES
	EXTENDED          = C.PCRE_EXTENDED
	EXTRA             = C.PCRE_EXTRA
	FIRSTLINE         = C.PCRE_FIRSTLINE
	JAVASCRIPT_COMPAT = C.PCRE_JAVASCRIPT_COMPAT
	MULTILINE         = C.PCRE_MULTILINE
	NO_AUTO_CAPTURE   = C.PCRE_NO_AUTO_CAPTURE
	UNGREEDY          = C.PCRE_UNGREEDY
	UTF8              = C.PCRE_UTF8
)

// Flags for Match functions
const (
	NOTBOL            = C.PCRE_NOTBOL
	NOTEOL            = C.PCRE_NOTEOL
	NOTEMPTY          = C.PCRE_NOTEMPTY
	NOTEMPTY_ATSTART  = C.PCRE_NOTEMPTY_ATSTART
	NO_START_OPTIMIZE = C.PCRE_NO_START_OPTIMIZE
	PARTIAL_HARD      = C.PCRE_PARTIAL_HARD
	PARTIAL_SOFT      = C.PCRE_PARTIAL_SOFT
)

// Flags for Study function
const (
	STUDY_JIT_COMPILE              = C.PCRE_STUDY_JIT_COMPILE
	STUDY_JIT_PARTIAL_SOFT_COMPILE = C.PCRE_STUDY_JIT_PARTIAL_SOFT_COMPILE
	STUDY_JIT_PARTIAL_HARD_COMPILE = C.PCRE_STUDY_JIT_PARTIAL_HARD_COMPILE
)

// Exec-time and get/set-time error codes
const (
	ERROR_NOMATCH        = C.PCRE_ERROR_NOMATCH
	ERROR_NULL           = C.PCRE_ERROR_NULL
	ERROR_BADOPTION      = C.PCRE_ERROR_BADOPTION
	ERROR_BADMAGIC       = C.PCRE_ERROR_BADMAGIC
	ERROR_UNKNOWN_OPCODE = C.PCRE_ERROR_UNKNOWN_OPCODE
	ERROR_UNKNOWN_NODE   = C.PCRE_ERROR_UNKNOWN_NODE
	ERROR_NOMEMORY       = C.PCRE_ERROR_NOMEMORY
	ERROR_NOSUBSTRING    = C.PCRE_ERROR_NOSUBSTRING
	ERROR_MATCHLIMIT     = C.PCRE_ERROR_MATCHLIMIT
	ERROR_CALLOUT        = C.PCRE_ERROR_CALLOUT
	ERROR_BADUTF8        = C.PCRE_ERROR_BADUTF8
	ERROR_BADUTF8_OFFSET = C.PCRE_ERROR_BADUTF8_OFFSET
	ERROR_PARTIAL        = C.PCRE_ERROR_PARTIAL
	ERROR_BADPARTIAL     = C.PCRE_ERROR_BADPARTIAL
	ERROR_RECURSIONLIMIT = C.PCRE_ERROR_RECURSIONLIMIT
	ERROR_INTERNAL       = C.PCRE_ERROR_INTERNAL
	ERROR_BADCOUNT       = C.PCRE_ERROR_BADCOUNT
	ERROR_JIT_STACKLIMIT = C.PCRE_ERROR_JIT_STACKLIMIT
)

// A reference to a compiled regular expression.
// Use Compile or MustCompile to create such objects.
type Regexp struct {
	ptr   []byte
	extra []byte
}

// Number of bytes in the compiled pattern
func pcresize(ptr *C.pcre) (size C.size_t) {
	C.pcre_fullinfo(ptr, nil, C.PCRE_INFO_SIZE, unsafe.Pointer(&size))
	return
}
func pcreJITsize(ptr *C.pcre, ext *C.pcre_extra) (size C.size_t) {
	C.pcre_fullinfo(ptr, ext, C.PCRE_INFO_JITSIZE, unsafe.Pointer(&size))
	return
}

// Number of capture groups
func pcregroups(ptr *C.pcre) (count C.int) {
	C.pcre_fullinfo(ptr, nil,
		C.PCRE_INFO_CAPTURECOUNT, unsafe.Pointer(&count))
	return
}

// Try to compile the pattern. If an error occurs, the second return
// value is non-nil.
func Compile(pattern string, flags int) (Regexp, error) {
	pattern1 := C.CString(pattern)
	defer C.free(unsafe.Pointer(pattern1))
	if clen := int(C.strlen(pattern1)); clen != len(pattern) {
		return Regexp{}, fmt.Errorf("%s (%d): %s",
			pattern,
			clen,
			"NUL byte in pattern",
		)
	}
	var errptr *C.char
	var erroffset C.int
	ptr := C.pcre_compile(pattern1, C.int(flags), &errptr, &erroffset, nil)
	if ptr == nil {
		return Regexp{}, fmt.Errorf("%s (%d): %s",
			pattern,
			int(erroffset),
			C.GoString(errptr),
		)
	}
	defer C.free(unsafe.Pointer(ptr))
	psize := pcresize(ptr)
	var re Regexp
	re.ptr = make([]byte, psize)
	C.memcpy(unsafe.Pointer(&re.ptr[0]), unsafe.Pointer(ptr), psize)
	return re, nil
}

// Compile pattern with jit compilation
func CompileJIT(ptr string, flags int) (Regexp, error) {
	re, errC := Compile(ptr, flags)
	if errC != nil {
		return Regexp{}, fmt.Errorf("Compile error: %s", errC)
	}
	errS := re.Study(C.PCRE_STUDY_JIT_COMPILE)
	if errS != nil {
		return re, fmt.Errorf("Study error: %s", errS)
	}
	return re, nil
}

// Compile the pattern. If compilation fails, panic.
func MustCompile(pattern string, flags int) (re Regexp) {
	re, err := Compile(pattern, flags)
	if err != nil {
		panic(err)
	}
	return
}

// CompileJIT the pattern. If compilation fails, panic.
func MustCompileJIT(pattern string, flags int) (re Regexp) {
	re, err := CompileJIT(pattern, flags)
	if err != nil {
		panic(err)
	}
	return
}

// Return the start and end of the first match.
func (re *Regexp) FindAllIndex(bytes []byte, flags int) (r [][]int) {
	m := re.Matcher(bytes, flags)
	offset := 0
	for m.Match(bytes[offset:], flags) {
		r = append(r, []int{offset + int(m.ovector[0]), offset + int(m.ovector[1])})
		offset += int(m.ovector[1])
	}
	return
}

// Return the start and end of the first match, or nil if no match.
// loc[0] is the start and loc[1] is the end.
func (re *Regexp) FindIndex(bytes []byte, flags int) []int {
	m := re.Matcher(bytes, flags)
	if m.Matches {
		return []int{int(m.ovector[0]), int(m.ovector[1])}
	}
	return nil
}

// Returns the number of capture groups in the compiled regexp pattern.
func (re Regexp) Groups() int {
	if re.ptr == nil {
		panic("Regexp.Groups: uninitialized")
	}
	return int(pcregroups((*C.pcre)(unsafe.Pointer(&re.ptr[0]))))
}

// Tries to match the speficied byte array slice to the current pattern.
// Returns true if the match succeeds.
func (r *Regexp) Match(subject []byte, flags int) bool {
	m := r.Matcher(subject, flags)
	return m.Matches
}

// Same as Match, but accept string as argument
func (r *Regexp) MatchString(subject string, flags int) bool {
	m := r.Matcher([]byte(subject), flags)
	return m.Matches
}

// Returns a new matcher object, with the byte array slice as a
// subject.
func (re Regexp) Matcher(subject []byte, flags int) (m *Matcher) {
	m = new(Matcher)
	m.Reset(re, subject, flags)
	return
}

// Returns a new matcher object, with the specified subject string.
func (re Regexp) MatcherString(subject string, flags int) (m *Matcher) {
	m = new(Matcher)
	m.ResetString(re, subject, flags)
	return
}

// Return a copy of a byte slice with pattern matches replaced by repl.
func (re Regexp) ReplaceAll(bytes, repl []byte, flags int) []byte {
	m := re.Matcher(bytes, 0)
	r := []byte{}
	for m.Match(bytes, flags) {
		r = append(append(r, bytes[:m.ovector[0]]...), repl...)
		bytes = bytes[m.ovector[1]:]
	}
	return append(r, bytes...)
}

// Same as ReplaceAll, but accept strings as arguments
func (re Regexp) ReplaceAllString(src, repl string, flags int) string {
	return string(re.ReplaceAll([]byte(src), []byte(repl), flags))
}

// Study regexp and add pcre_extra information to it, which gives huge
// speed boost when matching. If an error occurs, return value is
// non-nil. If flags = 0, don't study at all and return error.
// Studying can be quite a heavy optimization, but it's worth it.
func (re *Regexp) Study(flags int) error {
	if re.extra != nil {
		return fmt.Errorf("Regexp already optimized")
	}
	if flags == 0 {
		return fmt.Errorf("flag must be > 0")
	}
	var err *C.char
	extra := C.pcre_study((*C.pcre)(unsafe.Pointer(&re.ptr[0])), C.int(flags), &err)
	if err != nil {
		return fmt.Errorf(C.GoString(err))
	}
	defer C.free(unsafe.Pointer(extra))
	size := pcreJITsize((*C.pcre)(unsafe.Pointer(&re.ptr[0])), extra)
	if size > 0 {
		re.extra = make([]byte, size)
		C.memcpy(unsafe.Pointer(&re.extra[0]), unsafe.Pointer(extra), size)
		return nil
	} else {
		return fmt.Errorf(C.GoString(err))
	}
}

// Matcher objects provide a place for storing match results.
// They can be created by the Matcher and MatcherString functions,
// or they can be initialized with Reset or ResetString.
type Matcher struct {
	re       Regexp
	Groups   int
	ovector  []int32 // space for capture offsets, int32 is analogfor C.int type
	Matches  bool    // last match was successful
	Partial  bool    // was the last match a partial match?
	SubjectS string  // contain finded subject as string
	SubjectB []byte  // contain finded subject as []byte
}

// Tries to match the speficied byte array slice to the current
// pattern. Returns exec result.
// C docs http://www.pcre.org/original/doc/html/pcre_exec.html
func (m *Matcher) Exec(subject []byte, flags int) int {
	if m.re.ptr == nil {
		panic("Matcher.Match: uninitialized")
	}
	length := len(subject)
	m.SubjectS = string(subject)
	m.SubjectB = subject
	if length == 0 {
		subject = nullbyte // make first character adressable
	}
	subjectptr := (*C.char)(unsafe.Pointer(&subject[0]))
	return m.exec(subjectptr, length, flags)
}

// Same as Exec, but accept string as argument
func (m *Matcher) ExecString(subject string, flags int) int {
	return m.Exec([]byte(subject), flags)
}

func (m *Matcher) exec(subjectptr *C.char, length, flags int) int {
	var extra *C.pcre_extra
	if m.re.extra != nil {
		extra = (*C.pcre_extra)(unsafe.Pointer(&m.re.extra[0]))
	} else {
		extra = nil
	}
	rc := C.pcre_exec((*C.pcre)(unsafe.Pointer(&m.re.ptr[0])), extra,
		subjectptr, C.int(length), 0, C.int(flags),
		(*C.int)(unsafe.Pointer(&m.ovector[0])), C.int(len(m.ovector)))
	return int(rc)
}

// Returns the captured string with submatches of the last match
// (performed by Matcher, MatcherString, Reset, ResetString, Match,
// or MatchString). Group 0 is the part of the subject which matches
// the whole pattern; the first actual capture group is numbered 1.
// Capture groups which are not present return a nil slice.
func (m *Matcher) Extract() [][]byte {
	if m.Matches {
		captured_texts := make([][]byte, m.Groups+1)
		captured_texts[0] = m.SubjectB
		for i := 1; i < m.Groups+1; i++ {
			start := m.ovector[2*i]
			end := m.ovector[2*i+1]
			captured_text := m.SubjectB[start:end]
			captured_texts[i] = captured_text
		}
		return captured_texts
	} else {
		return nil
	}
}

// Same as Extract, but returns []string
func (m *Matcher) ExtractString() []string {
	if m.Matches {
		captured_texts := make([]string, m.Groups+1)
		captured_texts[0] = m.SubjectS
		for i := 1; i < m.Groups+1; i++ {
			start := m.ovector[2*i]
			end := m.ovector[2*i+1]
			captured_text := m.SubjectS[start:end]
			captured_texts[i] = captured_text
		}
		return captured_texts
	} else {
		return nil
	}
}

func (m *Matcher) init(re Regexp) {
	m.Matches = false
	if m.re.ptr != nil && &m.re.ptr[0] == &re.ptr[0] {
		// Skip group count extraction if the matcher has
		// already been initialized with the same regular
		// expression.
		return
	}
	m.re = re
	m.Groups = re.Groups()
	if ovectorlen := 3 * (1 + m.Groups); len(m.ovector) < ovectorlen {
		m.ovector = make([]int32, int32(ovectorlen))
	}
}

var nullbyte = []byte{0}

// Returns the numbered capture group of the last match (performed by
// Matcher, MatcherString, Reset, ResetString, Match, or MatchString).
// Group 0 is the part of the subject which matches the whole pattern;
// the first actual capture group is numbered 1. Capture groups which
// are not present return a nil slice.
func (m *Matcher) Group(group int) []byte {
	start := m.ovector[2*group]
	end := m.ovector[2*group+1]
	if start >= 0 {
		return m.SubjectB[start:end]
	}
	return nil
}

// Returns the numbered capture group positions of the last match
// (performed by Matcher, MatcherString, Reset, ResetString, Match,
// or MatchString). Group 0 is the part of the subject which matches
// the whole pattern; the first actual capture group is numbered 1.
// Capture groups which are not present return a nil slice.
func (m *Matcher) GroupIndices(group int) []int {
	start := m.ovector[2*group]
	end := m.ovector[2*group+1]
	if start >= 0 {
		return []int{int(start), int(end)}
	}
	return nil
}

// Same as Group, but returns string
func (m *Matcher) GroupString(group int) string {
	start := m.ovector[2*group]
	end := m.ovector[2*group+1]
	if start >= 0 {
		return m.SubjectS[start:end]
	}
	return ""
}

// Index returns the start and end of the first match, if a previous
// call to Matcher, MatcherString, Reset, ResetString, Match or
// MatchString succeeded. loc[0] is the start and loc[1] is the end.
func (m *Matcher) Index() []int {
	if !m.Matches {
		return nil
	}

	return []int{int(m.ovector[0]), int(m.ovector[1])}
}

// Tries to match the speficied byte array slice to the current
// pattern. Returns true if the match succeeds.
func (m *Matcher) Match(subject []byte, flags int) bool {
	rc := m.Exec(subject, flags)
	m.Matches, _ = checkMatch(rc)
	m.Partial = (rc == C.PCRE_ERROR_PARTIAL)
	return m.Matches
}

// Tries to match the speficied subject string to the current pattern.
// Returns true if the match succeeds.
func (m *Matcher) MatchString(subject string, flags int) bool {
	rc := m.ExecString(subject, flags)
	m.Matches, _ = checkMatch(rc)
	m.Partial = (rc == ERROR_PARTIAL)
	return m.Matches
}

func checkMatch(rc int) (bool, error) {
	switch {
	case rc >= 0 || rc == ERROR_PARTIAL:
		return true, nil
	case rc == ERROR_NOMATCH:
		return false, nil
	case rc == ERROR_NULL:
		return false, fmt.Errorf("pcre_exec: one or more variables passed to pcre_exec == NULL")
	case rc == ERROR_BADOPTION:
		return false, fmt.Errorf("pcre_exec: An unrecognized bit was set in the options argument")
	case rc == ERROR_BADMAGIC:
		return false, fmt.Errorf("pcre_exec: invalid option flag")
	case rc == ERROR_UNKNOWN_OPCODE:
		return false, fmt.Errorf("pcre_exec: an unknown item was encountered in the compiled pattern")
	case rc == ERROR_NOMEMORY:
		return false, fmt.Errorf("pcre_exec: match limit")
	case rc == ERROR_MATCHLIMIT:
		return false, fmt.Errorf("pcre_exec: backtracking (match) limit was reached")
	case rc == ERROR_BADUTF8:
		return false, fmt.Errorf("pcre_exec: string that contains an invalid UTF-8 byte sequence was passed as a subject")
	case rc == ERROR_RECURSIONLIMIT:
		return false, fmt.Errorf("pcre_exec: recursion limit")
	case rc == ERROR_JIT_STACKLIMIT:
		return false, fmt.Errorf("pcre_exec: error JIT stack limit")
	case rc == ERROR_INTERNAL:
		panic("pcre_exec: INTERNAL ERROR")
	case rc == ERROR_BADCOUNT:
		panic("pcre_exec: INTERNAL ERROR")
	}
	panic("unexepected return code from pcre_exec: " +
		strconv.Itoa(int(rc)))
}

func (m *Matcher) name2index(name string) (group int, err error) {
	if m.re.ptr == nil {
		err = fmt.Errorf("Matcher.Named: uninitialized")
		return
	}
	name1 := C.CString(name)
	defer C.free(unsafe.Pointer(name1))
	group = int(C.pcre_get_stringnumber(
		(*C.pcre)(unsafe.Pointer(&m.re.ptr[0])), name1))
	if group < 0 {
		err = fmt.Errorf("Matcher.Named: unknown name: " + name)
		return
	}
	return
}

// Returns the value of the named capture group. This is a nil slice
// if the capture group is not present. Panics if the name does not
// refer to a group.
func (m *Matcher) Named(group string) (g []byte, err error) {
	group_num, err := m.name2index(group)
	if err != nil {
		return
	}
	return m.Group(group_num), nil
}

// Returns true if the named capture group is present. Panics if the
// name does not refer to a group.
func (m *Matcher) NamedPresent(group string) (pres bool) {
	group_num, err := m.name2index(group)
	if err != nil {
		return false
	}
	return m.Present(group_num)
}

// Returns the value of the named capture group, or an empty string if
// the capture group is not present. Panics if the name does not
// refer to a group.
func (m *Matcher) NamedString(group string) (g string, err error) {
	group_num, err := m.name2index(group)
	if err != nil {
		return
	}
	return m.GroupString(group_num), nil
}

// Returns true if the numbered capture group is present in the last
// match (performed by Matcher, MatcherString, Reset, ResetString,
// Match, or MatchString). Group numbers start at 1. A capture group
// can be present and match the empty string.
func (m *Matcher) Present(group int) bool {
	return m.ovector[2*group] >= 0
}

// Switches the matcher object to the specified pattern and subject.
func (m *Matcher) Reset(re Regexp, subject []byte, flags int) {
	if re.ptr == nil {
		panic("Regexp.Matcher: uninitialized")
	}
	m.init(re)
	m.Match(subject, flags)
}

// Switches the matcher object to the specified pattern and subject
// string.
func (m *Matcher) ResetString(re Regexp, subject string, flags int) {
	if re.ptr == nil {
		panic("Regexp.Matcher: uninitialized")
	}
	m.init(re)
	m.MatchString(subject, flags)
}
