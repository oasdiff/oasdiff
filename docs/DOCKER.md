## Running oasdiff from docker
To run oasdiff from docker just replace the oasdiff command by `docker run --rm -t tufin/oasdiff` (or the image you prefer).  
Here are a few examples:

### Breaking changes with Docker
```bash
docker run --rm -t tufin/oasdiff breaking https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test1.yaml https://raw.githubusercontent.com/oasdiff/oasdiff/main/data/openapi-test3.yaml
```

### Comparing local files with Docker
```bash
docker run --rm -t -v $(pwd):/specs:ro -w /specs tufin/oasdiff changelog openapi-test1.yaml openapi-test3.yaml
```

Replace `$(pwd)` by the path that contains your specs.  
