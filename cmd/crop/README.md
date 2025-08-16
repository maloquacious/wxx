# crop

**⚠️ This command has been replaced by the `resize` command.**

The `crop` functionality is now available in the `resize` command using negative values for the parameters.

## Migration

Instead of using `crop`, use `resize` with negative values:

### Old crop usage:
```bash
crop -input map.wxx -output smaller.wxx -top 2 -left 1
```

### New resize equivalent:
```bash
resize -input map.wxx -output smaller.wxx -top -2 -left -1
```

## Benefits of using resize

The `resize` command provides:
- Both expansion and cropping in a single tool
- Better coordinate translation for features and labels
- More consistent parameter handling
- Enhanced validation and error reporting

See the [resize command documentation](../resize/README.md) for full details.
