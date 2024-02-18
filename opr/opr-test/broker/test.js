'use strict';

const { Gateway, Wallets } = require('fabric-network');
const fs = require('fs');
const path = require('path');
const YAML = require('yamljs');

const Kafka = require('node-rdkafka');

async function main() {
    const a = {a : 1, b : 2, c : "test", d : "test2"};
    var c = a.c;
    if (a.c != undefined) {
        a.c = undefined;
    }
    console.log(a.f);
    console.log(a);
    a.c = c;
    console.log(a);
}

main();
