# Design Notes

## Examples

Can download a package with `pip download there` to get `there-0.0.12-py2.py3-none-any.whl`.

Can upload to our GitLab with `curl` or `twine`

```bash
curl -v --request POST \
  --form 'content=@there-0.0.12-py2.py3-none-any.whl' \
  --form 'name=there' \
  --form 'version=0.0.12' \
  --user $ACT3_USERNAME:$ACT3_TOKEN \
  "https://git.act3-ace.com/api/v4/projects/796/packages/pypi?requires_python=3.7"
```

```bash
twine upload --repository-url "https://git.act3-ace.com/api/v4/projects/796/packages/pypi" -u $ACT3_USERNAME -p $ACT3_TOKEN there-0.0.12-py2.py3-none-any.whl --verbose
```

## Docs

There is no accepted PEP for uploading a python distribution file.  This [PEP-694](https://peps.python.org/pep-0694/) is trying to make a standard and overcome deficiencies in the current API used by PyPI.org.  An old and incomplete one does exist [PEP-243](https://peps.python.org/pep-0243/).

## Example Twine Upload Form Data

```txt
--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="name"

there
--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="version"

0.0.12
--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="filetype"

bdist_wheel
--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="pyversion"

py2.py3
--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="metadata_version"

2.1
--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="summary"

Print current file and line number
--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="home_page"


--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="author"


--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="author_email"

Matthias Bussonnier <bussonniermatthias@gmail.com>
--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="maintainer"


--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="maintainer_email"


--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="license"


--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="description"

OMMITED DESCRIPTION


--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="keywords"


--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="classifiers"

License :: OSI Approved :: MIT License
--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="download_url"


--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="comment"


--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="sha256_digest"

de8d6eaf343f48635640b7d092c13a265b955e52f85d423f410cb57bd8436afe
--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="project_urls"

Home, https://github.com/Carreau/there
--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="requires_python"


--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="description_content_type"

text/markdown
--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="md5_digest"

e14443174ebd71d5893327824d3f1e70
--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="blake2_256_digest"

d431c2aa184c8579c478e12b4b4bbfef50a209a0d03c55b23040a2820db6eeff
--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name=":action"

file_upload
--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="protocol_version"

1
--4f042191a20e4a9d9179bf52aec96376
Content-Disposition: form-data; name="content"; filename="there-0.0.12-py2.py3-none-any.whl"
Content-Type: application/octet-stream
```
