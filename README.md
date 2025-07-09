# nimbus

## TODO:

- [x] frpc setup
- [x] api to control vms
- [ ] testing newly provisioned machines accessible
- [ ] MINECRAFT SERVER
- [ ] frontend website
- [ ] pool to instantly provision

## bugs:
- [ ] Incorrect shutdown, something to do with signals. I don't think this is fixable, since it is likely caused by interrupts being sent to firecracker process as well as sectionleader. Workaround for now is shutdown-all endpoint, and can delete veth network interfaces to fix.
