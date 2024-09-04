## Things that have to be done before the project is ready for production

### Before merge
- Update @kyvejs/sdk and @kyvejs/types in protocol/core to the latest version
  - Hint: This are the only dependencies that need to still be maintained in the old repo once the RDK is released
- Sync the code in protocol/core with the core in the kyvejs repo
- Make changes on the chain (if necessary) to support the new RDK
- Merge [PR](https://github.com/KYVENetwork/kyve-rdk/pull/1) to main

### After merge
- Fix TODOs in tools/kysor
  - Especially changing the repo name in `tools/kysor/cmd/types/const.go` to `github.com/KYVENetwork/kyve-rdk`. 
    This points currently to a repo that already uses the new CI/CD pipeline. After the merge this is not necessary anymore.
- Setup release-please to automatically create releases
  - Set the initial release-please commit in `.release-please-manifest.json` 
  to the commit that was created by the PR (Checkout [replease-please docs](https://github.com/googleapis/release-please/blob/main/docs/cli.md#bootstrapping))

### Optional
- Make the python template in `tools/kystrap/templates/python` production ready
- Make the typescript template in `tools/kystrap/templates/typescript` production ready

Hint: The go template is already production ready (maybe it needs some minor changes)

### When ready for production
- Update all the docs in this repo
- Update the KYVE docs and explain how to use the new RDK

### Open issues
There are some issues that contain more or less the same information as this document.
- https://linear.app/kyve/issue/TECH-1692/upgrade-pool-configs
- https://linear.app/kyve/issue/TECH-1746/rdk-before-merge
- https://linear.app/kyve/issue/TECH-1696/update-kystrap-template-for-typescript
- https://linear.app/kyve/issue/TECH-1697/update-kystrap-template-for-python
- https://linear.app/kyve/issue/TECH-819/docs-for-rdk