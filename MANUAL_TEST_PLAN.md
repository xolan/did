# Manual Integration Test Plan - Corrupted Storage Recovery

## Overview
This document describes the manual integration tests for the corrupted storage recovery feature. These tests verify that the `did` tool gracefully handles corrupted JSONL entries and provides appropriate warnings to users.

## Prerequisites
1. Build the binary: `just build` or `go build -o did .`
2. Backup your existing entries file if you have one: `cp ~/.config/did/entries.jsonl ~/.config/did/entries.jsonl.backup`

## Test Scenarios

### Test 1: Valid Entries Only (Baseline)
**Purpose:** Verify normal operation with no corrupted entries

**Setup:**
```bash
# Create a clean test file
cat > ~/.config/did/entries.jsonl << 'EOF'
{"Timestamp":"2026-01-02T10:00:00Z","Description":"Feature A","DurationMinutes":120,"RawInput":"Feature A for 2h"}
{"Timestamp":"2026-01-02T14:00:00Z","Description":"Bug fix B","DurationMinutes":45,"RawInput":"Bug fix B for 45m"}
EOF
```

**Test Commands:**
```bash
./did                    # List today's entries
./did validate          # Validate storage
```

**Expected Results:**
- `did` should list 2 entries with no warnings
- `did validate` should show:
  - Total lines: 2
  - Valid entries: 2
  - Corrupted entries: 0
  - Status: ✓ Storage file is healthy

---

### Test 2: Malformed JSON at Beginning
**Purpose:** Verify handling of corrupted entry at start of file

**Setup:**
```bash
cat > ~/.config/did/entries.jsonl << 'EOF'
{this is not valid json}
{"Timestamp":"2026-01-02T10:00:00Z","Description":"Feature A","DurationMinutes":120,"RawInput":"Feature A for 2h"}
{"Timestamp":"2026-01-02T14:00:00Z","Description":"Bug fix B","DurationMinutes":45,"RawInput":"Bug fix B for 45m"}
EOF
```

**Test Commands:**
```bash
./did                    # List today's entries
./did validate          # Validate storage
```

**Expected Results:**
- `did` should:
  - Display warning to stderr: "Warning: Found 1 corrupted line(s) in storage file:"
  - Show: "  Line 1: {this is not valid json} (error: ...)"
  - Display blank line after warnings
  - List 2 valid entries (Feature A, Bug fix B)
  - Calculate total duration correctly (2h 45m)

- `did validate` should show:
  - Storage file path
  - Total lines: 3
  - Valid entries: 2
  - Corrupted entries: 1
  - Corrupted lines section with: "  Line 1: {this is not valid json} (error: ...)"
  - Status: ⚠ Storage file has 1 corrupted line(s) (to stderr)

---

### Test 3: Malformed JSON in Middle
**Purpose:** Verify handling of corrupted entry between valid entries

**Setup:**
```bash
cat > ~/.config/did/entries.jsonl << 'EOF'
{"Timestamp":"2026-01-02T10:00:00Z","Description":"Feature A","DurationMinutes":120,"RawInput":"Feature A for 2h"}
incomplete json line without closing brace
{"Timestamp":"2026-01-02T14:00:00Z","Description":"Bug fix B","DurationMinutes":45,"RawInput":"Bug fix B for 45m"}
EOF
```

**Test Commands:**
```bash
./did
./did validate
```

**Expected Results:**
- Both commands should handle corruption gracefully
- Valid entries before and after corrupted line should load
- Warning should show line 2 is corrupted
- Total duration should only include valid entries (2h 45m)

---

### Test 4: Multiple Corrupted Lines
**Purpose:** Verify handling of multiple corrupted entries

**Setup:**
```bash
cat > ~/.config/did/entries.jsonl << 'EOF'
{"Timestamp":"2026-01-02T10:00:00Z","Description":"Feature A","DurationMinutes":120,"RawInput":"Feature A for 2h"}
{corrupted line 1}
{"Timestamp":"2026-01-02T12:00:00Z","Description":"Feature C","DurationMinutes":30,"RawInput":"Feature C for 30m"}
not even json at all
{"Timestamp":"2026-01-02T14:00:00Z","Description":"Bug fix B","DurationMinutes":45,"RawInput":"Bug fix B for 45m"}
EOF
```

**Test Commands:**
```bash
./did
./did validate
```

**Expected Results:**
- Warning message should say: "Found 2 corrupted line(s)"
- Both corrupted lines (2 and 4) should be listed
- 3 valid entries should load correctly
- Total duration should be 3h 15m (120 + 30 + 45 minutes)

---

### Test 5: Truncated Long Corrupted Line
**Purpose:** Verify that very long corrupted lines are truncated in warnings

**Setup:**
```bash
cat > ~/.config/did/entries.jsonl << 'EOF'
{"Timestamp":"2026-01-02T10:00:00Z","Description":"Feature A","DurationMinutes":120,"RawInput":"Feature A for 2h"}
{"this is a very long corrupted line that should be truncated because it exceeds fifty characters in length and contains invalid JSON"}
{"Timestamp":"2026-01-02T14:00:00Z","Description":"Bug fix B","DurationMinutes":45,"RawInput":"Bug fix B for 45m"}
EOF
```

**Test Commands:**
```bash
./did
./did validate
```

**Expected Results:**
- Corrupted line should be truncated to 50 characters with "..." appended
- Warning should show: "  Line 2: {"this is a very long corrupted line that shoul... (error: ...)"
- Valid entries should load normally

---

### Test 6: Empty Lines
**Purpose:** Verify handling of empty or whitespace-only lines

**Setup:**
```bash
cat > ~/.config/did/entries.jsonl << 'EOF'
{"Timestamp":"2026-01-02T10:00:00Z","Description":"Feature A","DurationMinutes":120,"RawInput":"Feature A for 2h"}

{"Timestamp":"2026-01-02T14:00:00Z","Description":"Bug fix B","DurationMinutes":45,"RawInput":"Bug fix B for 45m"}
EOF
```

**Test Commands:**
```bash
./did
./did validate
```

**Expected Results:**
- Empty line (line 2) should be reported as corrupted
- Valid entries should load
- Warning should mention "unexpected end of JSON input" or similar

---

### Test 7: All Lines Corrupted
**Purpose:** Verify behavior when entire file is corrupted

**Setup:**
```bash
cat > ~/.config/did/entries.jsonl << 'EOF'
this is not json
neither is this
or this one
EOF
```

**Test Commands:**
```bash
./did
./did validate
```

**Expected Results:**
- Warning should say: "Found 3 corrupted line(s)"
- All 3 lines should be listed as corrupted
- `did` should show: "No entries found for today"
- No crash or error exit

---

### Test 8: Non-existent File
**Purpose:** Verify graceful handling when storage file doesn't exist

**Setup:**
```bash
rm -f ~/.config/did/entries.jsonl
```

**Test Commands:**
```bash
./did
./did validate
```

**Expected Results:**
- `did` should show: "No entries found for today"
- `did validate` should show:
  - Total lines: 0
  - Valid entries: 0
  - Corrupted entries: 0
  - Status: ✓ Storage file is healthy
- No errors or crashes

---

### Test 9: Date Filtering with Corrupted Lines
**Purpose:** Verify date filtering commands work with corrupted entries

**Setup:**
```bash
cat > ~/.config/did/entries.jsonl << 'EOF'
{"Timestamp":"2026-01-01T10:00:00Z","Description":"Yesterday work","DurationMinutes":60,"RawInput":"Yesterday work for 1h"}
{corrupted line}
{"Timestamp":"2026-01-02T10:00:00Z","Description":"Today work","DurationMinutes":120,"RawInput":"Today work for 2h"}
EOF
```

**Test Commands:**
```bash
./did      # Today
./did y    # Yesterday
./did w    # This week
./did lw   # Last week
```

**Expected Results:**
- All commands should show the same corruption warning (line 2)
- Each command should correctly filter entries by date
- No duplicate warnings

---

### Test 10: Adding New Entry to Corrupted File
**Purpose:** Verify new entries can be added despite existing corruption

**Setup:**
```bash
cat > ~/.config/did/entries.jsonl << 'EOF'
{"Timestamp":"2026-01-02T10:00:00Z","Description":"Feature A","DurationMinutes":120,"RawInput":"Feature A for 2h"}
{corrupted}
EOF
```

**Test Commands:**
```bash
./did new feature for 1h
./did
```

**Expected Results:**
- New entry should be added successfully
- When listing entries:
  - Warning about line 2 corruption
  - Both valid entries (Feature A and new feature) should display
  - Total duration should be 3h

---

## Test Execution Checklist

- [ ] Test 1: Valid Entries Only (Baseline)
- [ ] Test 2: Malformed JSON at Beginning
- [ ] Test 3: Malformed JSON in Middle
- [ ] Test 4: Multiple Corrupted Lines
- [ ] Test 5: Truncated Long Corrupted Line
- [ ] Test 6: Empty Lines
- [ ] Test 7: All Lines Corrupted
- [ ] Test 8: Non-existent File
- [ ] Test 9: Date Filtering with Corrupted Lines
- [ ] Test 10: Adding New Entry to Corrupted File

## Acceptance Criteria Verification

After running all tests, verify:

- [x] Malformed JSON lines are skipped with a warning to stderr ✓
- [x] Valid entries before and after corrupted lines are still loaded ✓
- [x] Warning includes the line number and content of corrupted entries ✓
- [x] `did validate` command shows storage file health status ✓
- [x] Corrupted entries don't cause crashes or data loss for valid entries ✓

## Cleanup

After testing, restore your original entries file:
```bash
mv ~/.config/did/entries.jsonl.backup ~/.config/did/entries.jsonl
```

## Notes for Tester

- All warnings should go to stderr, not stdout
- Valid entry listing should go to stdout
- Line numbers in warnings should be 1-indexed (first line is line 1)
- Content in warnings should be truncated at 50 characters with "..." if longer
- The tool should never crash due to corrupted entries
- Exit codes should be 0 for successful operations, even with corrupted entries present
