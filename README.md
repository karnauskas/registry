Docker Registry
============

A multiuser docker registry to be used behind https proxies

Installation
------------
```
docker build -t docker-registry .
```

Usage
-----
```
docker run -p 5000:5000 docker-registry
```

License
-------
[GPLv3][gpl3.0]

[gpl3.0]: https://www.gnu.org/licenses/gpl-3.0.txt

Contributing
------------
Please follow the [Open Code of Conduct][code-of-conduct].

[code-of-conduct]: http://todogroup.org/opencodeofconduct

To make sure your pull request will be accepted, please open an issue in the issue tracker before starting work where we can talk with you to make sure a feature or bug fix is going in a direction where it will benifit everybody.

TODO
----
- Add DELETE Method to Manifest for image removal
- Add Json Error Messages That Docker Will Understand
