'use strict';

const { Wallets } = require('fabric-network');
const FabricCAServices = require('fabric-ca-client');
const fs = require('fs');
const path = require('path');
const YAML = require('yamljs');

exports.main = async function main(orcid, email) {
    const user = orcid;
    const passwd = Math.random().toString(36)
    //const argv = process.argv;
    //const user = argv[2];
    //const passwd = argv[3];
    const msp = 'JISTAP';
    const aff = 'jistap';
    const userType = 'reviewer';
    //const orcid = user;
    //const email = user;
    //const orcid = argv[7];
    //const email = argv[8];
    try {
        // load the network configuration
        const ccpPath = path.resolve('/', 'root', 'fabric', 'script', 'connection.yaml');
        const ccp = YAML.parse(fs.readFileSync(ccpPath, 'utf8'));

        // Create a new CA client for interacting with the CA.
        const caURL = ccp.certificateAuthorities.ca.url;
        const ca = new FabricCAServices(caURL);

        // Create a new file system based wallet for managing identities.
        const walletPath = path.join(process.cwd(), 'wallet');
        const wallet = await Wallets.newFileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        // Check to see if we've already enrolled the user.
        const userIdentity = await wallet.get(user);
        if (userIdentity) {
            console.log('An identity for the user ${user} already exists in the wallet');
            return;
        }

        // Check to see if we've already enrolled the admin user.
        const adminIdentity = await wallet.get('admin');
        if (!adminIdentity) {
            console.log('An identity for the admin user "admin" does not exist in the wallet');
            console.log('Run the enrollAdmin.js application before retrying');
            return;
        }

        // build a user object for authenticating with the CA
        const provider = wallet.getProviderRegistry().getProvider(adminIdentity.type);
        const adminUser = await provider.getUserContext(adminIdentity, 'admin');

        // Register the user, enroll the user, and import the new identity into the wallet.
        const secret = await ca.register({
            affiliation: aff,
            enrollmentID: user,
            enrollmentSecret: passwd,
            role: 'client',
            attrs: [ { name: "userType", value: userType, ecert: true }, { name: "identifier", value: orcid, ecert: true }, { name: "email", value: email, ecert: true } ],
        }, adminUser);
        const enrollment = await ca.enroll({
            enrollmentID: user,
            enrollmentSecret: passwd,
            attr_reqs: [{ name: "userType", optional: false }, { name: "identifier", optional: false }, { name: "email", optional: true }]
        });
        const x509Identity = {
            credentials: {
                certificate: enrollment.certificate,
                privateKey: enrollment.key.toBytes(),
            },
            mspId: msp,
            type: 'X.509',
        };
        await wallet.put(user, x509Identity);
        console.log(`Successfully registered and enrolled admin user ${user} and imported it into the wallet`);

    } catch (error) {
        console.error(`Failed to register user ${user}: ${error}`);
        process.exit(1);
    }
}

//main();
