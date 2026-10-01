import json, pathlib, shutil
base = pathlib.Path("/private/tmp/go-independent-review-kolmdafa")
source = base / "candidate"
changes = {
    "in-place-publication": ('\t// Keep the temporary file in the destination directory', '\tif true { return os.WriteFile(path, data, 0o644) }\n\t// Keep the temporary file in the destination directory'),
    "accept-blank-fields": ('if strings.TrimSpace(item.Code) == "" || strings.TrimSpace(item.Label) == "" {', 'if false && (strings.TrimSpace(item.Code) == "" || strings.TrimSpace(item.Label) == "") {'),
    "ignore-late-read-error": ('if err != nil {\n\t\treturn fmt.Errorf("read snapshot: %w", err)', 'if err != nil && len(data) == 0 {\n\t\treturn fmt.Errorf("read snapshot: %w", err)'),
    "omit-response-close": ('\tdefer resp.Body.Close()', '\t// Deliberately omit response-body closure.'),
    "accept-trailing-json": ('json.Unmarshal(data, &items)', 'json.NewDecoder(strings.NewReader(string(data))).Decode(&items)'),
}
manifest = []
for name, (old, new) in changes.items():
    dest = base / "mutations" / name
    shutil.copytree(source, dest)
    path = dest / "mirror.go"
    text = path.read_text()
    assert text.count(old) == 1, name
    path.write_text(text.replace(old, new, 1))
    manifest.append(dict(name=name, path=str(path), old=old, new=new))
(base / "output" / "mutations.json").write_text(json.dumps(manifest, indent=2) + "\n")
print("Created five disposable single-defect mutations; original and candidate were not modified.")

