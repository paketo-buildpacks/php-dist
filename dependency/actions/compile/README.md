Running compilation locally:

1. Build the build environment:
```
docker build --platform <os>/<arch> -t compilation -f <target>.Dockerfile dependency/actions/compile
```

2. Make the output directory:
```
mkdir <output dir>
```

3. Run compilation and use a volume mount to access it:
```
docker run --platform <os>/<arch> -v <output dir>:$PWD --rm compilation --version <version> --outputDir $PWD --target <target> --os <os> --arch <arch>
```

Notes:
- Use <os>=linux and <arch>=amd64 or arm64.
- If you omit --platform/--os/--arch, defaults are linux/x64.
