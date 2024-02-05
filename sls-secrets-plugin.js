'use strict';

const {SecretsManagerClient, GetSecretValueCommand} = require("@aws-sdk/client-secrets-manager");
const {fromIni} = require("@aws-sdk/credential-providers");
const fs = require("fs");

const _secretsFile = './.secrets.json'

class SlsSecretsPlugin {
    constructor(serverless, options = {}) {
        this.options = options;
        this.serverless = serverless;
        const pkgHooks = {
            // "before:package:createDeploymentArtifacts": this.beforePackage.bind(this),
            "before:package:initialize": this.beforePackage.bind(this),
            "before:deploy:function:initialize": this.beforePackage.bind(this),
            // For serverless-offline plugin
            "before:offline:start:init": this.offlineStart.bind(this),
            // For invoke local
            "before:invoke:local:invoke": this.beforePackage.bind(this),
        };

        const cleanupHooks = {
            "after:package:finalize": this.cleanupPackage.bind(this),
            "after:deploy:finalize": this.cleanupPackage.bind(this),
            "after:deploy:function:deploy": this.cleanupPackage.bind(this),
            // For serverless-offline plugin
            "before:offline:start:end": this.offlineStop.bind(this),
            // For invoke local
            "after:invoke:local:invoke": this.cleanupPackage.bind(this),
        };

        this.hooks = {...pkgHooks, ...cleanupHooks};
    }

    async loadSecrets(profile) {
        const config = this.serverless.configurationInput
        const params = config.custom
        profile = profile || config.provider.profile

        console.log(`Loading secrets from ${JSON.stringify(params.secrets_arn)}`)

        const cfg = {
            region: this.serverless.configurationInput.provider.region,
        }

        if (process.env['CI_ENV'] !== 'yes') {
            cfg.credentials = fromIni({profile: profile})
        }
        const client = new SecretsManagerClient(cfg);

        let merged = {}
        for (let i = 0; i < params.secrets_arn.length; i++) {
            const value = await client.send(new GetSecretValueCommand({
                SecretId: params.secrets_arn[i]
            }));

            console.log(`Loaded secret ${params.secrets_arn[i]}`)
            merged = {...merged, ...JSON.parse(value.SecretString)}
        }

        return merged
    }

    async offlineStart() {
        const secrets = await this.loadSecrets('tracker')
        fs.writeFileSync(_secretsFile, JSON.stringify(secrets));
    }

    async offlineStop() {
        if (fs.existsSync(_secretsFile)) fs.unlinkSync(_secretsFile);
    }


    async beforePackage() {
        const secrets = await this.loadSecrets()
        // Before deploy
        fs.writeFileSync(_secretsFile, JSON.stringify(secrets));
    }

    cleanupPackage() {
        if (fs.existsSync(_secretsFile)) fs.unlinkSync(_secretsFile);
    }
}

module.exports = SlsSecretsPlugin;