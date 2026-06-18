# Changelog

## [0.3.0](https://github.com/omar-kada/air-compose/compare/v0.2.0...v0.3.0) (2026-06-18)


### Features

* add unread notification indicator ([#170](https://github.com/omar-kada/air-compose/issues/170)) ([e75ba7a](https://github.com/omar-kada/air-compose/commit/e75ba7a219610b16a06041eac671c0b38fa3b05d))
* deployment on config updated ([#156](https://github.com/omar-kada/air-compose/issues/156)) ([1e51346](https://github.com/omar-kada/air-compose/commit/1e513468e5c91c99cbb0813081c11b5b0fa7a43e))
* handle events in client using websockets ([#166](https://github.com/omar-kada/air-compose/issues/166)) ([9f242b0](https://github.com/omar-kada/air-compose/commit/9f242b040f38e46a7b7599e194eb205a962a8bbc))
* redeploy on unhealthy ([#144](https://github.com/omar-kada/air-compose/issues/144)) ([b8fac76](https://github.com/omar-kada/air-compose/commit/b8fac760c77b1463eb7896e61e9b2989bc8cb7c1))
* unhealthy stacks detection and notification ([#124](https://github.com/omar-kada/air-compose/issues/124)) ([6bb52e6](https://github.com/omar-kada/air-compose/commit/6bb52e6da6ce0549c912b4c6a9d2b749fb26ff83))


### Bug Fixes

* double override inside env file ([#115](https://github.com/omar-kada/air-compose/issues/115)) ([cdc33c1](https://github.com/omar-kada/air-compose/commit/cdc33c144d5591acdfe4e9d20e866c0f1455ce39))
* refactor health check and state ([#137](https://github.com/omar-kada/air-compose/issues/137)) ([0fe65d4](https://github.com/omar-kada/air-compose/commit/0fe65d49a5e6f31d4b9ea84e14f4523a9a5155f5))
* stop serving directories in spa handler ([#127](https://github.com/omar-kada/air-compose/issues/127)) ([b049789](https://github.com/omar-kada/air-compose/commit/b049789a713fb2b52843eb70681a8f3577854628))

## [0.2.0](https://github.com/omar-kada/air-compose/compare/v0.1.2...v0.2.0) (2026-05-18)


### Features

* add http insecure mode ([#73](https://github.com/omar-kada/air-compose/issues/73)) ([2e554c8](https://github.com/omar-kada/air-compose/commit/2e554c8afd13d83306c76f3dc316ae6870502c9c))
* add onboarding screen ([#71](https://github.com/omar-kada/air-compose/issues/71)) ([403a27f](https://github.com/omar-kada/air-compose/commit/403a27f18986815522dc036e7b403389c193a315))
* add repo info to deployment details page ([#84](https://github.com/omar-kada/air-compose/issues/84)) ([483b1ee](https://github.com/omar-kada/air-compose/commit/483b1ee82901878a8326126a0e02f334c70c2bca))
* add test connection in onboarding form ([#77](https://github.com/omar-kada/air-compose/issues/77)) ([94c0d9e](https://github.com/omar-kada/air-compose/commit/94c0d9ebf63df7cc59deb3a6e835c5305d0d0f11))
* add timeline component in deployment event log ([#79](https://github.com/omar-kada/air-compose/issues/79)) ([abfd18e](https://github.com/omar-kada/air-compose/commit/abfd18e6bb3a83ee8a63b2879622deeaecff8ecc))
* adjust auth token security based on request ([#114](https://github.com/omar-kada/air-compose/issues/114)) ([e80686d](https://github.com/omar-kada/air-compose/commit/e80686d8998edf7e8e465b07f8a0872da71eb45d))
* oidc integration api ([#81](https://github.com/omar-kada/air-compose/issues/81)) ([f708406](https://github.com/omar-kada/air-compose/commit/f708406065b10bf28343ee81f0a25707e5ac9eae))
* oidc integration client ([#82](https://github.com/omar-kada/air-compose/issues/82)) ([d5e0d7a](https://github.com/omar-kada/air-compose/commit/d5e0d7acba81607167829f94dd955fff32ce3e3c))


### Bug Fixes

* mobile UI spacing & scroll enhancements ([#78](https://github.com/omar-kada/air-compose/issues/78)) ([5ed9c72](https://github.com/omar-kada/air-compose/commit/5ed9c72b00e548a9fc9fc202a464efa1fc84ea31))
* organize settings in sub objects ([#83](https://github.com/omar-kada/air-compose/issues/83)) ([fc895af](https://github.com/omar-kada/air-compose/commit/fc895af2a1d0557793e65b5b7ace8b02c6d158f4))

## [0.1.2](https://github.com/omar-kada/air-compose/compare/v0.1.1...v0.1.2) (2026-03-18)


### Bug Fixes

* Set Secure flag to false for cookies for testing ([bdb4556](https://github.com/omar-kada/air-compose/commit/bdb4556d556ff3c2db5a153b2c357a7769dc93fa))

## [0.1.1](https://github.com/omar-kada/air-compose/compare/v0.1.0...v0.1.1) (2026-03-04)


### Bug Fixes

* issue when changing repo url, diff should work immediatly ([#68](https://github.com/omar-kada/air-compose/issues/68)) ([e3d37fe](https://github.com/omar-kada/air-compose/commit/e3d37fe80fd8c7d6b4e75ebfc7457a5eb7341471))

## [0.1.0](https://github.com/omar-kada/air-compose/compare/v0.0.1...v0.1.0) (2026-03-01)


### Features

* auto-sync using cron period ([c74b1f8](https://github.com/omar-kada/air-compose/commit/c74b1f8a0242e5a0cf46365da4c577f6507c7714))
* display and edit local configuration ([c74b1f8](https://github.com/omar-kada/air-compose/commit/c74b1f8a0242e5a0cf46365da4c577f6507c7714))
* display and edit settings ([c74b1f8](https://github.com/omar-kada/air-compose/commit/c74b1f8a0242e5a0cf46365da4c577f6507c7714))
* display deployment events ([c74b1f8](https://github.com/omar-kada/air-compose/commit/c74b1f8a0242e5a0cf46365da4c577f6507c7714))
* display diff with git repo ([c74b1f8](https://github.com/omar-kada/air-compose/commit/c74b1f8a0242e5a0cf46365da4c577f6507c7714))
* display stacks status ([c74b1f8](https://github.com/omar-kada/air-compose/commit/c74b1f8a0242e5a0cf46365da4c577f6507c7714))
* feature flags for settings and configuration update ([c74b1f8](https://github.com/omar-kada/air-compose/commit/c74b1f8a0242e5a0cf46365da4c577f6507c7714))
* load stacks from git repo and deploy them ([c74b1f8](https://github.com/omar-kada/air-compose/commit/c74b1f8a0242e5a0cf46365da4c577f6507c7714))
* send notifications using shoutrrr ([c74b1f8](https://github.com/omar-kada/air-compose/commit/c74b1f8a0242e5a0cf46365da4c577f6507c7714))
* single user authentication ([c74b1f8](https://github.com/omar-kada/air-compose/commit/c74b1f8a0242e5a0cf46365da4c577f6507c7714))
