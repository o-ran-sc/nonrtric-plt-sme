/*
   ========================LICENSE_START=================================
   O-RAN-SC
   %%
   Copyright (C) 2023: Nordix Foundation
   %%
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
   ========================LICENSE_END===================================
*/

function printResources(resources) {
    let res = Object.values(resources);
    let out = `<p class="lead">Resources:</p> `;
    res.forEach((r) => {
        out += `
            <p>
            <strong>CommType:</strong> ${r.commType}
            <strong>CustOpName:</strong> ${r.custOpName} 
            <strong>ResourceName:</strong> ${r.resourceName}
            <strong>Uri:</strong> ${r.uri} 
            <strong>Description:</strong> ${r.description}
            <strong>Operations:</strong> ${Object.values(r.operations)}
            </p>
        `;
    });
    return out;
}

function printCustomOperations(custOperations) {
    let operations = Object.values(custOperations);
    let out = `<p class="lead"> Custom Operations:</p> `;
    operations.forEach((o) => {
        out += `
            <p>
            <strong>CommType:</strong> ${o.commType}
            <strong>CustOpName:</strong> ${o.custOpName} 
            <strong>Description:</strong> ${o.description}
            <strong>Operations:</strong> ${Object.values(o.operations)}
            </p>
        `;
    });
    return out;
}

function printVersions(versions) {
    let vers = Object.values(versions);
    let out = `<p class="lead">Versions:</p>`;
    vers.forEach((v) => {
        out += `
            <p>
            <strong>ApiVersion:</strong> ${v.apiVersion} 
            ${printCustomOperations(v.custOperations)}
            ${printResources(v.resources)}
            </p>
        `;
    });
    return out;
}

function printInterfaceDescription(description) {
    let interfaceDescriptions = Object.values(description);
    let out = `<p class="lead">Interface Description:</p>`;
    interfaceDescriptions.forEach((d) => {
        out += `
            <p>
            <strong>Ipv4Addr:</strong> ${d.ipv4Addr} 
            <strong>Ipv6Addr:</strong> ${d.ipv6Addr} 
            <strong>Port:</strong> ${d.port} 
            <strong>SecurityMethods:</strong> ${Object.values(d.securityMethods)}
            </p>
        `;
    });
    return out;
}

function printAefProfiles(aefProfiles){
    let out = "";
    let index = 0;
    aefProfiles.forEach((aef) => {
      out += `
         <tr data-bs-toggle="collapse"  data-bs-target="#r${index}">
            <td>${aef.aefId}</td>
            <td>${aef.aefLocation}</td>
            <td>${aef.domainName}</td>
            <td>${aef.protocol}</td>
            <td>${Object.values(aef.securityMethods)}</td>
         </tr>
         <tr class="collapse accordion-collapse" id="r${index}" data-bs-parent=".table">
            <td colspan="5"> 
                <div id="demo1">
                    ${printInterfaceDescription(aef.interfaceDescriptions)} 
                    ${printVersions(aef.versions)}
                </div> 
            </td>
        </tr>
      `;
      index++;
    }); 
    return out;
}