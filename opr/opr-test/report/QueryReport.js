/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Gateway, Wallets } = require('fabric-network');
const path = require('path');
const fs = require('fs');
const YAML = require('yamljs');


exports.main = async function main(user, contractKey) {
    //const argv = process.argv;
    //const user = argv[2];
    //const organization = argv[3];
    //const paperID = argv[4];
    //const round = argv[5];
    //const contractKey = "".concat('\u0000', '\u0000', organization, '\u0000', paperID, '\u0000', round, '\u0000');

    try {
        // load the network configuration
        const ccpPath = path.resolve('/', 'root', 'fabric', 'script', 'connection.yaml');
        const ccp = YAML.parse(fs.readFileSync(ccpPath, 'utf8'));

        // Create a new file system based wallet for managing identities.
        const walletPath = path.join(process.cwd(), 'wallet');
        const wallet = await Wallets.newFileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        // Check to see if we've already enrolled the user.
        const identity = await wallet.get(user);
        if (!identity) {
            console.log(`An identity for the user ${user} does not exist in the wallet`);
            console.log('Run the registerUser.js application before retrying');
            return;
        }

        // Create a new gateway for connecting to our peer node.
        const gateway = new Gateway();
        await gateway.connect(ccp, { wallet, identity: user, discovery: { enabled: false, asLocalhost: true } });

        // Get the network (channel) our contract is deployed to.
        const network = await gateway.getNetwork('boprs');

        // Get the contract from the network.
        const contract = network.getContract('report');

        // Evaluate the specified transaction.
        const result = await contract.evaluateTransaction('QueryReport', contractKey);
        console.log(`Transaction has been evaluated, result is: ${JSON.stringify(JSON.parse(result.toString().replace(/\\u0000/g, '_')), null, 2)}`);
        return result

    } catch (error) {
        console.error(`Failed to evaluate transaction: ${error}`);
        return 0
    }
}

//main();
