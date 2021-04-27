# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

### [1.5.1](https://github.com/alexfalkowski/go-health/compare/v1.5.0...v1.5.1) (2021-04-27)

## [1.5.0](https://github.com/alexfalkowski/go-health/compare/v1.4.1...v1.5.0) (2021-04-05)


### Features

* rename packages ([#27](https://github.com/alexfalkowski/go-health/issues/27)) ([4abba53](https://github.com/alexfalkowski/go-health/commit/4abba53eca7887ef6205266d38492c87b4d15004))

### [1.4.1](https://github.com/alexfalkowski/go-health/compare/v1.4.0...v1.4.1) (2021-04-04)

## [1.4.0](https://github.com/alexfalkowski/go-health/compare/v1.3.0...v1.4.0) (2021-04-03)


### Features

* update go to version 1.16 ([#24](https://github.com/alexfalkowski/go-health/issues/24)) ([494df76](https://github.com/alexfalkowski/go-health/commit/494df76909d74f28a38f39b6e7c282b0f839ad7a))

## [1.3.0](https://github.com/alexfalkowski/go-health/compare/v1.2.0...v1.3.0) (2020-11-04)


### Features

* break up to interfaces ([#23](https://github.com/alexfalkowski/go-health/issues/23)) ([a37a5c5](https://github.com/alexfalkowski/go-health/commit/a37a5c572a84f54f293e65e70ace96cbf826d8c0))

## [1.2.0](https://github.com/alexfalkowski/go-health/compare/v1.1.0...v1.2.0) (2020-11-02)


### Features

* **tcp:** add ability to pass dialer ([#22](https://github.com/alexfalkowski/go-health/issues/22)) ([cb79b53](https://github.com/alexfalkowski/go-health/commit/cb79b534053a2ab2189969fa2fe06f18b72b2cd0))

## 1.1.0 (2020-10-31)


### Features

* **checker:** add tcp checker ([8aa2989](https://github.com/alexfalkowski/go-health/commit/8aa2989530c8c38ac1799b5e6e58ba3d10f8e8fc))
* **ci:** use new image ([06ce5cb](https://github.com/alexfalkowski/go-health/commit/06ce5cb972735c1cd03d96a6cdd3f7e2a2d8ca7b))
* **db:** check sql ([51a9215](https://github.com/alexfalkowski/go-health/commit/51a92152ff628ca7100659b7715f9da592b44fc2))
* **http:** add probe ([5b46b87](https://github.com/alexfalkowski/go-health/commit/5b46b878508e23335b7ba5c1e0c9f0b2bb8afd4a))
* **observer:** allow a subscription to be observed ([31ca88f](https://github.com/alexfalkowski/go-health/commit/31ca88fc1a7caeec78d3800b83170fcd527e712c))
* **srv:** default with 10 secs ([9137b45](https://github.com/alexfalkowski/go-health/commit/9137b45d2b9d29b7f55c4bcd89ffec9022877f06))
* **srv:** pass in registrations ([a8c5258](https://github.com/alexfalkowski/go-health/commit/a8c5258af299b995f0a0624cf534c20eb91ce0fb))
* **stringer:** add string function ([3f0ff6c](https://github.com/alexfalkowski/go-health/commit/3f0ff6c058f7620a7f1f83ac0939e939e93ffa29))
* **stringer:** change to quotes ([a2cd3d5](https://github.com/alexfalkowski/go-health/commit/a2cd3d5fe4b20ffed9838c90441c283c8d9d4d1b))
* add editor config ([b627803](https://github.com/alexfalkowski/go-health/commit/b627803bb3cc440f8b16e7c37fc5e5784a903a76))
* add git ignore ([f686839](https://github.com/alexfalkowski/go-health/commit/f686839c5ea8cbd90c3367592392a6c22fa6d5fa))
* add linter config ([b294cf4](https://github.com/alexfalkowski/go-health/commit/b294cf457c1a6cc26f6bf7287a825bbf6eecd9a6))
* add makefile ([5bd35a5](https://github.com/alexfalkowski/go-health/commit/5bd35a5e8ea395d19afd2b04e6dcf550fa3529c2))
* go mod init github.com/alexfalkowski/go-health ([75290e2](https://github.com/alexfalkowski/go-health/commit/75290e2ddb916350f5ef5923c6199522f0b028f7))
* send tick ([f0eee1c](https://github.com/alexfalkowski/go-health/commit/f0eee1cf946cf86cf2469ab0ab654db2c1729879))


### Bug Fixes

* linting from the new linter ([50df0e7](https://github.com/alexfalkowski/go-health/commit/50df0e73d3bfe7639cfc81c1dce64811f49110e0))
* **http:** pass tripper ([4413d47](https://github.com/alexfalkowski/go-health/commit/4413d47028201a3c121385a4db9c72d8384568a5))
* **http:** return nil ([478ee01](https://github.com/alexfalkowski/go-health/commit/478ee01d521365b72339ce53a3861e8fc3d0a53c))
* **srv:** only subscribe to exiting registries ([b534295](https://github.com/alexfalkowski/go-health/commit/b534295027350707542f5133346b448d5299aa24))
* **tcp:** close connection ([d8f35bb](https://github.com/alexfalkowski/go-health/commit/d8f35bb6c4f7a7d9d307646abbe616041e31d60b))
* buffer channel ([14b8e56](https://github.com/alexfalkowski/go-health/commit/14b8e5644323330c3b608dcff64a83840f94661b))
* close the channel ([b514335](https://github.com/alexfalkowski/go-health/commit/b514335cb7ae4d795eb5d07745b786ac946175d7))
* handle stoping ([d519ad0](https://github.com/alexfalkowski/go-health/commit/d519ad0f3dc76148d9a7ea2e597882375db66117))
* mantain a status for starting and stopping ([9f8d4c5](https://github.com/alexfalkowski/go-health/commit/9f8d4c546458277eb07e9b235a8ef2a02c0c6eae))
* wait for cancellation ([7ccc20f](https://github.com/alexfalkowski/go-health/commit/7ccc20ff29f165ecf6469dbe987e182233b388f5))
