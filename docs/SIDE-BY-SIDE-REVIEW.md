# Side-by-Side Review in Your Browser (`--open`)
Both `breaking` and `changelog` accept an `--open` flag. After printing the usual terminal output, the CLI uploads the comparison to [oasdiff.com](https://www.oasdiff.com) and opens the rendered side-by-side review in your browser, with the diff anchored at each change site. The approve/reject buttons are shown but disabled; they are enabled only with [oasdiff Pro](https://www.oasdiff.com/pricing).

```
oasdiff changelog HEAD~1:openapi.yaml HEAD:openapi.yaml --open
```

The comparison is encrypted on your machine before upload, with a one-time key that lives only in the review link (the part after the `#`) and is never sent to our servers, so what we store we cannot read. There is no sign-in and no account.

The resulting URL works for 7 days. Paste it into a PR comment or Slack and reviewers can open the review without installing the CLI themselves. Anyone you give the full link to can open it (the part after the `#` is the decryption key), so treat it like a secret.
