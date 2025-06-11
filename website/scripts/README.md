# Website Build Scripts

Scripts for building the Osprey website.

## Scripts

### `generate-docs.sh`
Generates reference documentation from Osprey compiler examples.

- Builds compiler if needed
- Extracts symbols from `.osp` files using `osprey --docs`
- Generates `stdlib.md`, `types.md`, `functions.md`

**Usage:**
```bash
./scripts/generate-docs.sh
```

### `copy-spec.js`
Copies language specification from compiler to website source.

## Manual Documentation Generation

```bash
cd compiler
./osprey path/to/file.osp --docs
``` 