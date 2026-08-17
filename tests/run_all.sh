#!/bin/sh
# run_all.sh – Runs all tmpl8 test cases and reports PASS/FAIL per test.
# Usage: sh tests/run_all.sh  (from repo root, or just ./run_all.sh from tests/)
# Prerequisites: tmpl8 on $PATH, or the binary at ../bin/tmpl8 relative to this script.

# ---------------------------------------------------------------------------
# Resolve the tmpl8 binary
# ---------------------------------------------------------------------------
if [ -f "$(dirname "$0")/../bin/tmpl8" ]; then
    TMPL8="$(cd "$(dirname "$0")/.." && pwd)/bin/tmpl8"
elif command -v tmpl8 >/dev/null 2>&1; then
    TMPL8=$(which tmpl8)
else
    echo "ERROR: tmpl8 not found on \$PATH or at ../bin/tmpl8" >&2
    exit 1
fi

echo "Using tmpl8 binary: $TMPL8"

DIR="$(cd "$(dirname "$0")" && pwd)"

PASS_COUNT=0
FAIL_COUNT=0
TOTAL=8

# ---------------------------------------------------------------------------
# Helper: run_test NAME EXPECTED_FILE CMD [ARGS...]
#   Runs CMD, diffs stdout against EXPECTED_FILE.
#   Prints PASS or FAIL (with diff) and updates counters.
# ---------------------------------------------------------------------------
run_test() {
    NAME="$1"
    EXPECTED="$2"
    shift 2
    TMPFILE=$(mktemp)
    "$@" > "$TMPFILE" 2>/dev/null
    if diff -u "$EXPECTED" "$TMPFILE" >/dev/null 2>&1; then
        echo "PASS: $NAME"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        echo "FAIL: $NAME"
        diff -u "$EXPECTED" "$TMPFILE" || true
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
    rm -f "$TMPFILE"
}

# ---------------------------------------------------------------------------
# Helper: run_test_filtered NAME EXPECTED_FILE FILTER_PATTERN CMD [ARGS...]
#   Like run_test but strips lines matching FILTER_PATTERN from both
#   the actual output and the expected file before diffing.
# ---------------------------------------------------------------------------
run_test_filtered() {
    NAME="$1"
    EXPECTED="$2"
    FILTER="$3"
    shift 3
    TMPFILE=$(mktemp)
    TMPACTUAL=$(mktemp)
    TMPEXPECTED=$(mktemp)
    "$@" > "$TMPFILE" 2>/dev/null
    grep -v "$FILTER" "$EXPECTED" > "$TMPEXPECTED" || true
    grep -v "$FILTER" "$TMPFILE"   > "$TMPACTUAL"   || true
    if diff -u "$TMPEXPECTED" "$TMPACTUAL" >/dev/null 2>&1; then
        echo "PASS: $NAME"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        echo "FAIL: $NAME"
        diff -u "$TMPEXPECTED" "$TMPACTUAL" || true
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
    rm -f "$TMPFILE" "$TMPACTUAL" "$TMPEXPECTED"
}

# ---------------------------------------------------------------------------
# Helper: run_fail_test NAME CMD [ARGS...]
#   Passes if CMD exits with a non-zero exit code.
# ---------------------------------------------------------------------------
run_fail_test() {
    NAME="$1"
    shift
    if ! "$@" >/dev/null 2>&1; then
        echo "PASS: $NAME"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        echo "FAIL: $NAME (expected non-zero exit code, got zero)"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
}

# ---------------------------------------------------------------------------
# Test 01 – Inline data + inline template (smoke test)
# ---------------------------------------------------------------------------
cd "$DIR"
run_test "01 – inline smoke" \
    "$DIR/01_inline_smoke/expected.txt" \
    "$TMPL8" -i '?name: world' '?Hello, {{ .name }}!'

# ---------------------------------------------------------------------------
# Test 02 – JSON input → plain-text output
# ---------------------------------------------------------------------------
cd "$DIR/02_json_to_text"
run_test "02 – JSON to text" \
    expected.txt \
    "$TMPL8" -i input.json template.tmpl

# ---------------------------------------------------------------------------
# Test 03 – YAML input → YAML output
# ---------------------------------------------------------------------------
cd "$DIR/03_yaml_to_yaml"
run_test "03 – YAML to YAML" \
    expected.yaml \
    "$TMPL8" -i input.yaml template.tmpl

# ---------------------------------------------------------------------------
# Test 04 – Empty input with -z flag (version line filtered)
# ---------------------------------------------------------------------------
cd "$DIR/04_no_input"
run_test_filtered "04 – no input (-z)" \
    expected.txt \
    "^tmpl8 version:" \
    "$TMPL8" -z template.tmpl

# ---------------------------------------------------------------------------
# Test 05 – fail guard on missing input (must exit non-zero)
# ---------------------------------------------------------------------------
cd "$DIR/05_fail_on_invalid"
run_fail_test "05 – fail guard" \
    "$TMPL8" -z template.tmpl

# ---------------------------------------------------------------------------
# Test 06 – readfile + format conversions (toYaml, fromJson)
# ---------------------------------------------------------------------------
cd "$DIR/06_readfile_and_conversions"
run_test "06 – readfile and conversions" \
    expected.yaml \
    "$TMPL8" -i input.json template.tmpl

# ---------------------------------------------------------------------------
# Test 07 – Multiple templates chained together
# ---------------------------------------------------------------------------
cd "$DIR/07_chained_templates"
run_test "07 – chained templates" \
    expected.txt \
    "$TMPL8" -i input.yaml header.tmpl body.tmpl footer.tmpl

# ---------------------------------------------------------------------------
# Test 08 – Kubernetes ConfigMap generator (end-to-end)
# ---------------------------------------------------------------------------
cd "$DIR/08_k8s_configmap"
run_test "08 – K8s ConfigMap generator" \
    expected.yaml \
    "$TMPL8" -i tree.json kustomization.tmpl

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "${PASS_COUNT}/${TOTAL} tests passed"

if [ "$FAIL_COUNT" -gt 0 ]; then
    exit 1
fi
exit 0
