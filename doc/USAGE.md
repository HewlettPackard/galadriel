# Galadriel Server CLI
The galadriel server CLI contains the functionality to:
- Register a new `member`
    - ```bash
        bin/galadriel-server create member -t <trust-domain>
        ```
- Register a new `relationship`
    - ```bash
        bin/galadriel-server create relationship -a <trust-domain-A> -b <trust-domain-B>
        ```
- Generate a new `token`
    - ```bash
        bin/galadriel-server generate token -t <trust-domain>
        ```
- List all `members` and `relationships`
    - ```bash
        bin/galadriel-server list members
        ```
    - ```bash
        bin/galadriel-server list relationships
        ```

# Galadriel Harvester CLI
The galadriel Harvester CLI contains the functionality to run the harvester while attaching it to the Galadriel Server instance, based on the token used as a argument:

```bash
bin/galadriel-harvester run -t <ACCESS_TOKEN>
```