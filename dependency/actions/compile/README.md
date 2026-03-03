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
- <target> can be: jammy or noble
- <os>: linux
- <arch>: amd64 or arm64
- If you omit --platform/--os/--arch, defaults are linux/x64.

Example for PHP 8.5.2 on noble/arm64:
```
docker build --platform linux/arm64 -t compilation -f noble.Dockerfile dependency/actions/compile
docker run --platform linux/arm64 -v ~/php-build:/home --rm compilation --version 8.5.2 --outputDir /home --target noble --os linux --arch arm64
```

Example for PHP 8.5.2 on jammy/amd64:
```
docker build --platform linux/amd64 -t compilation -f jammy.Dockerfile dependency/actions/compile
docker run --platform linux/amd64 -v ~/php-build:/home --rm compilation --version 8.5.2 --outputDir /home --target jammy --os linux --arch amd64
```
