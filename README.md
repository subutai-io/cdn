# Gorjun
Gorjun is a golang replacement for Kurjun project. Kurjun is Subutai's CDN software.

## Vagrant

Use vagrant to create a multi-machine development and testing environment in an internal network. The master vm builds, configures and runs Gorjun. The cdn\[1-n] nodes are NGINX CDN cache nodes.

> vagrant up
> vagrant ssh master
> ping cdn1.local
> ping cdn2.local
