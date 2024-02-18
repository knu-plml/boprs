/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const FabricCAServices = require('fabric-ca-client');
const { Wallets } = require('fabric-network');
const fs = require('fs');
const path = require('path');
const YAML = require('yamljs');

async function main() {
    try {
        // load the network configuration
        const ccpPath = path.resolve('/', 'root', 'fabric', 'script', 'connection.yaml');
        const ccp = YAML.parse(fs.readFileSync(ccpPath, 'utf8'));
	const argv = process.argv;
	const user = argv[2];
	const pass = argv[3];

        // Create a new CA client for interacting with the CA.
        const caInfo = ccp.certificateAuthorities.ca;
        const caTLSCACerts = caInfo.tlsCACerts.path;
        const ca = new FabricCAServices(caInfo.url, { trustedRoots: caTLSCACerts, verify: false });

        // Create a new file system based wallet for managing identities.
        const walletPath = path.join(process.cwd(), 'wallet');
        const wallet = await Wallets.newFileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        // Check to see if we've already enrolled the admin user.
        const identity = await wallet.get(user);
        if (identity) {
            console.log('An identity for the admin user '+user+' already exists in the wallet');
            return;
        }

        // Enroll the admin user, and import the new identity into the wallet.
        const enrollment = await ca.enroll({
            //enrollmentID: 'testadmin', enrollmentSecret: 'testadminpw' });
            enrollmentID: user,
            enrollmentSecret: pass
    });
        const x509Identity = {
            credentials: {
                certificate: enrollment.certificate,
                privateKey: enrollment.key.toBytes(),
            },
            mspId: 'JISTAP',
            type: 'X.509',
        };
        await wallet.put(user, x509Identity);
        console.log('Successfully enrolled admin user '+user+' and imported it into the wallet');

    } catch (error) {
        console.error(`Failed to enroll admin user ${user}: ${error}`);
        process.exit(1);
    }
}

main();
