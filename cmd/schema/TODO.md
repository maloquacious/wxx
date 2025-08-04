# TODO: wxxschema Go Struct Generation

## High Priority Tasks

### ✅ Analysis Complete
- [x] **Analyze existing schema inference** - Understand how the current program works

### 🚧 Core Structure Generation

- [ ] **Implement generateGoStructs function**
  - Replace the current `generateHierarchy` function with one that outputs valid Go code
  - Add `_t` suffix to all struct names (e.g., `Map_t`, `Configuration_t`)
  - Generate proper struct field declarations with correct capitalization
  - Handle element names that need transformation (e.g., XML element "map" → Go struct "Map_t")

- [ ] **Add proper XML struct tags**
  - Include `xml:"elementName"` tags for all fields
  - Handle attribute vs element distinction with `xml:"attrName,attr"` for attributes
  - Support omitempty where appropriate
  - Consider xmlns namespace handling if needed

- [ ] **Ensure named nested structs**
  - Child elements must be typed as named structs (e.g., `Configuration Configuration_t`)
  - No anonymous struct definitions in generated code
  - Proper forward declarations if needed for circular references

- [ ] **Test generated structs**
  - Create test that unmarshals original XML using generated structs
  - Verify round-trip marshal/unmarshal preserves data
  - Test with both XML schema versions (1.0 and 1.1)

## Medium Priority Tasks

### 🔧 Enhanced Functionality

- [ ] **Create interface generator (if needed)**
  - Generate interfaces with `_i` suffix when beneficial
  - Consider interfaces for polymorphic XML elements
  - Document when interfaces are generated vs structs

- [ ] **Implement type inference**
  - Analyze XML content to determine field types beyond string
  - Support int, bool, float64 types based on content patterns
  - Handle optional vs required fields
  - Consider pointer types for optional elements

- [ ] **Handle repeated elements**
  - Detect when XML elements appear multiple times
  - Generate slice types `[]Element_t` for collections
  - Distinguish between optional single elements and arrays
  - Handle mixed content scenarios

### 🛠 Code Quality & Structure

- [ ] **Improve Element struct**
  - Add field to track if element can repeat (for slice generation)
  - Store inferred type information
  - Track element vs attribute distinction more explicitly

## Low Priority Tasks

### 📝 Polish & Usability

- [ ] **Generate proper package header**
  - Add package declaration
  - Include necessary imports (encoding/xml, etc.)
  - Add generation timestamp and source file comments
  - Include copyright notice if needed

- [ ] **Add command line flags**
  - `--input` flag to specify input WXX file
  - `--output` flag to specify output Go file
  - `--package` flag to set package name
  - `--help` and `--version` flags

### 🔍 Future Enhancements

- [ ] **Schema validation**
  - Validate generated structs compile successfully
  - Check for Go naming conflicts
  - Warn about potential issues (reserved keywords, etc.)

- [ ] **Multiple file generation**
  - Split large schemas into multiple Go files
  - Generate separate files per major XML element
  - Handle cross-file struct references

- [ ] **Documentation generation**
  - Add Go doc comments to generated structs
  - Include XML element descriptions if available
  - Generate usage examples

## Implementation Notes

### Current Code Structure
- `Element` struct represents XML schema nodes
- `inferSchema()` builds the element tree from XML
- `generateHierarchy()` outputs human-readable structure

### Required Changes
- Modify `generateHierarchy()` or create new `generateGoStructs()`
- Add type system to `Element` struct
- Handle Go naming conventions and reserved words
- Support multiple output formats (console, file)

### Testing Strategy
- Use existing test WXX files
- Generate structs and compile them
- Test XML unmarshaling with generated types
- Compare with manually created struct definitions
