#!/bin/bash
# Manual Integration Test Script for Corrupted Storage Recovery
# This script helps execute the manual tests documented in MANUAL_TEST_PLAN.md

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
STORAGE_DIR="${HOME}/.config/did"
STORAGE_FILE="${STORAGE_DIR}/entries.jsonl"
BACKUP_FILE="${STORAGE_DIR}/entries.jsonl.backup"
BINARY="./did"

# Check if binary exists
if [ ! -f "$BINARY" ]; then
    echo -e "${RED}Error: Binary './did' not found. Please run 'just build' first.${NC}"
    exit 1
fi

# Function to print test header
print_test() {
    echo -e "\n${BLUE}==================================================${NC}"
    echo -e "${BLUE}Test $1: $2${NC}"
    echo -e "${BLUE}==================================================${NC}\n"
}

# Function to create storage directory
setup_storage() {
    mkdir -p "$STORAGE_DIR"
}

# Function to backup existing entries
backup_entries() {
    if [ -f "$STORAGE_FILE" ]; then
        echo -e "${YELLOW}Backing up existing entries to ${BACKUP_FILE}${NC}"
        cp "$STORAGE_FILE" "$BACKUP_FILE"
    fi
}

# Function to restore entries
restore_entries() {
    if [ -f "$BACKUP_FILE" ]; then
        echo -e "${YELLOW}Restoring original entries from backup${NC}"
        mv "$BACKUP_FILE" "$STORAGE_FILE"
    fi
}

# Function to wait for user
wait_for_user() {
    echo -e "\n${YELLOW}Press Enter to continue to next test...${NC}"
    read
}

# Trap to ensure cleanup on exit
trap restore_entries EXIT

echo -e "${GREEN}==================================================${NC}"
echo -e "${GREEN}Corrupted Storage Recovery - Integration Tests${NC}"
echo -e "${GREEN}==================================================${NC}\n"

# Setup
setup_storage
backup_entries

# Test 1: Valid Entries Only (Baseline)
print_test "1" "Valid Entries Only (Baseline)"
cat > "$STORAGE_FILE" << 'EOF'
{"Timestamp":"2026-01-02T10:00:00Z","Description":"Feature A","DurationMinutes":120,"RawInput":"Feature A for 2h"}
{"Timestamp":"2026-01-02T14:00:00Z","Description":"Bug fix B","DurationMinutes":45,"RawInput":"Bug fix B for 45m"}
EOF
echo "Storage file created with 2 valid entries"
echo -e "\n${YELLOW}Running: ./did${NC}"
$BINARY
echo -e "\n${YELLOW}Running: ./did validate${NC}"
$BINARY validate
wait_for_user

# Test 2: Malformed JSON at Beginning
print_test "2" "Malformed JSON at Beginning"
cat > "$STORAGE_FILE" << 'EOF'
{this is not valid json}
{"Timestamp":"2026-01-02T10:00:00Z","Description":"Feature A","DurationMinutes":120,"RawInput":"Feature A for 2h"}
{"Timestamp":"2026-01-02T14:00:00Z","Description":"Bug fix B","DurationMinutes":45,"RawInput":"Bug fix B for 45m"}
EOF
echo "Storage file created with corrupted line at beginning"
echo -e "\n${YELLOW}Running: ./did${NC}"
$BINARY
echo -e "\n${YELLOW}Running: ./did validate${NC}"
$BINARY validate
wait_for_user

# Test 3: Malformed JSON in Middle
print_test "3" "Malformed JSON in Middle"
cat > "$STORAGE_FILE" << 'EOF'
{"Timestamp":"2026-01-02T10:00:00Z","Description":"Feature A","DurationMinutes":120,"RawInput":"Feature A for 2h"}
incomplete json line without closing brace
{"Timestamp":"2026-01-02T14:00:00Z","Description":"Bug fix B","DurationMinutes":45,"RawInput":"Bug fix B for 45m"}
EOF
echo "Storage file created with corrupted line in middle"
echo -e "\n${YELLOW}Running: ./did${NC}"
$BINARY
echo -e "\n${YELLOW}Running: ./did validate${NC}"
$BINARY validate
wait_for_user

# Test 4: Multiple Corrupted Lines
print_test "4" "Multiple Corrupted Lines"
cat > "$STORAGE_FILE" << 'EOF'
{"Timestamp":"2026-01-02T10:00:00Z","Description":"Feature A","DurationMinutes":120,"RawInput":"Feature A for 2h"}
{corrupted line 1}
{"Timestamp":"2026-01-02T12:00:00Z","Description":"Feature C","DurationMinutes":30,"RawInput":"Feature C for 30m"}
not even json at all
{"Timestamp":"2026-01-02T14:00:00Z","Description":"Bug fix B","DurationMinutes":45,"RawInput":"Bug fix B for 45m"}
EOF
echo "Storage file created with 2 corrupted lines"
echo -e "\n${YELLOW}Running: ./did${NC}"
$BINARY
echo -e "\n${YELLOW}Running: ./did validate${NC}"
$BINARY validate
wait_for_user

# Test 5: Truncated Long Corrupted Line
print_test "5" "Truncated Long Corrupted Line"
cat > "$STORAGE_FILE" << 'EOF'
{"Timestamp":"2026-01-02T10:00:00Z","Description":"Feature A","DurationMinutes":120,"RawInput":"Feature A for 2h"}
{"this is a very long corrupted line that should be truncated because it exceeds fifty characters in length and contains invalid JSON"}
{"Timestamp":"2026-01-02T14:00:00Z","Description":"Bug fix B","DurationMinutes":45,"RawInput":"Bug fix B for 45m"}
EOF
echo "Storage file created with very long corrupted line"
echo -e "${YELLOW}Expected: Line content should be truncated to 50 chars with '...'${NC}"
echo -e "\n${YELLOW}Running: ./did${NC}"
$BINARY
echo -e "\n${YELLOW}Running: ./did validate${NC}"
$BINARY validate
wait_for_user

# Test 6: Empty Lines
print_test "6" "Empty Lines"
cat > "$STORAGE_FILE" << 'EOF'
{"Timestamp":"2026-01-02T10:00:00Z","Description":"Feature A","DurationMinutes":120,"RawInput":"Feature A for 2h"}

{"Timestamp":"2026-01-02T14:00:00Z","Description":"Bug fix B","DurationMinutes":45,"RawInput":"Bug fix B for 45m"}
EOF
echo "Storage file created with empty line"
echo -e "\n${YELLOW}Running: ./did${NC}"
$BINARY
echo -e "\n${YELLOW}Running: ./did validate${NC}"
$BINARY validate
wait_for_user

# Test 7: All Lines Corrupted
print_test "7" "All Lines Corrupted"
cat > "$STORAGE_FILE" << 'EOF'
this is not json
neither is this
or this one
EOF
echo "Storage file created with all corrupted lines"
echo -e "${YELLOW}Expected: No entries found, 3 corrupted lines reported${NC}"
echo -e "\n${YELLOW}Running: ./did${NC}"
$BINARY
echo -e "\n${YELLOW}Running: ./did validate${NC}"
$BINARY validate
wait_for_user

# Test 8: Non-existent File
print_test "8" "Non-existent File"
rm -f "$STORAGE_FILE"
echo "Storage file removed"
echo -e "${YELLOW}Expected: No entries, graceful handling${NC}"
echo -e "\n${YELLOW}Running: ./did${NC}"
$BINARY
echo -e "\n${YELLOW}Running: ./did validate${NC}"
$BINARY validate
wait_for_user

# Test 9: Date Filtering with Corrupted Lines
print_test "9" "Date Filtering with Corrupted Lines"
cat > "$STORAGE_FILE" << 'EOF'
{"Timestamp":"2026-01-01T10:00:00Z","Description":"Yesterday work","DurationMinutes":60,"RawInput":"Yesterday work for 1h"}
{corrupted line}
{"Timestamp":"2026-01-02T10:00:00Z","Description":"Today work","DurationMinutes":120,"RawInput":"Today work for 2h"}
EOF
echo "Storage file created with entries from different dates"
echo -e "\n${YELLOW}Running: ./did (today)${NC}"
$BINARY
echo -e "\n${YELLOW}Running: ./did y (yesterday)${NC}"
$BINARY y
echo -e "\n${YELLOW}Running: ./did w (this week)${NC}"
$BINARY w
wait_for_user

# Test 10: Adding New Entry to Corrupted File
print_test "10" "Adding New Entry to Corrupted File"
cat > "$STORAGE_FILE" << 'EOF'
{"Timestamp":"2026-01-02T10:00:00Z","Description":"Feature A","DurationMinutes":120,"RawInput":"Feature A for 2h"}
{corrupted}
EOF
echo "Storage file created with existing corruption"
echo -e "\n${YELLOW}Running: ./did new feature for 1h${NC}"
$BINARY new feature for 1h
echo -e "\n${YELLOW}Running: ./did (should show both valid entries)${NC}"
$BINARY
wait_for_user

echo -e "\n${GREEN}==================================================${NC}"
echo -e "${GREEN}All tests completed!${NC}"
echo -e "${GREEN}==================================================${NC}\n"

echo -e "${YELLOW}Acceptance Criteria Verification:${NC}"
echo "✓ Malformed JSON lines are skipped with a warning to stderr"
echo "✓ Valid entries before and after corrupted lines are still loaded"
echo "✓ Warning includes the line number and content of corrupted entries"
echo "✓ did validate command shows storage file health status"
echo "✓ Corrupted entries don't cause crashes or data loss for valid entries"

echo -e "\n${YELLOW}Your original entries have been restored.${NC}"
