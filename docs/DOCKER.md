# Running oasdiff from docker
To run oasdiff from docker just replace the oasdiff command by `docker run --rm -t tufin/oasdiff` (or the image you prefer).  
Here are a few examples:

## Breaking changes with Docker
```bash
docker run --rm -t tufin/oasdiff breaking https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

## Comparing local files with Docker
```bash
docker run --rm -t -v $(pwd):/specs:ro -w /specs tufin/oasdiff changelog openapi-test1.yaml openapi-test3.yaml
```

Replace `$(pwd)` by the path that contains your specs.  

## Running as a non-root user
The container runs as root. The actions in [oasdiff-action](https://github.com/oasdiff/oasdiff-action) build on this image and need root, so the image keeps it.

Pass `--user` to run as someone else:
```bash
docker run --user $(id -u):$(id -g) --rm -t -v $(pwd):/specs:ro -w /specs tufin/oasdiff changelog openapi-test1.yaml openapi-test3.yaml
```

Loading specs from [git revisions](GIT-REVISION.md) works as any user, because the image trusts a mounted repository whatever uid owns it. `--fetch` is the exception: it writes to `.git`, so it needs a user that can write to the mount. `$(id -u):$(id -g)` can, another uid cannot.
