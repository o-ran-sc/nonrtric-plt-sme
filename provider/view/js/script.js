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
const isObject = (value) => typeof value === "object" && value !== null

function checkValue(value){
    return (isObject(value) ? value : "");
}

function printResources(resources) {
    let res = Object.values(checkValue(resources));
    let out = `<p class="lead">Resources:</p><ul>`;
    res.forEach((r) => {
        out += `<li>
            <p>
            <strong>CommType:</strong> ${r.commType}
            <strong>CustOpName:</strong> ${r.custOpName}
            <strong>ResourceName:</strong> ${r.resourceName}
            <strong>Uri:</strong> ${r.uri}
            <strong>Description:</strong> ${r.description}
            <strong>Operations:</strong> ${Object.values(checkValue(r.operations))}
            </p>
            </li>`;
    });
    out += `</ul>`;
    return out;
}

function printCustomOperations(custOperations) {
    let operations = Object.values(checkValue(custOperations));
    let out = `<p class="lead"> Custom Operations:</p><ul>`;
    operations.forEach((o) => {
        out += `<li>
            <p>
            <strong>CommType:</strong> ${o.commType}
            <strong>CustOpName:</strong> ${o.custOpName}
            <strong>Description:</strong> ${o.description}
            <strong>Operations:</strong> ${Object.values(checkValue(o.operations))}
            </p>
            </li>`;
    });
    out += `</ul>`;
    return out;
}

function printVersions(versions) {
    let vers = Object.values(checkValue(versions));
    let out = `<p class="lead">Versions:</p><ul>`;
    vers.forEach((v) => {
        out += `<li>
            <p>
            <strong>ApiVersion:</strong> ${v.apiVersion}
            ${printCustomOperations(v.custOperations)}
            ${printResources(v.resources)}
            </p>
            </li>`;
    });
    out += `</ul>`;
    return out;
}

function printInterfaceDescription(description) {
    let interfaceDescriptions = Object.values(checkValue(description));
    let out = `<p class="lead">Interface Description:</p><ul>`;
    interfaceDescriptions.forEach((d) => {
        out += `<li>
            <p>
            <strong>Ipv4Addr:</strong> ${d.ipv4Addr}
            <strong>Ipv6Addr:</strong> ${d.ipv6Addr}
            <strong>Port:</strong> ${d.port}
            <strong>SecurityMethods:</strong> ${Object.values(checkValue(d.securityMethods))}
            </p>
            </li>`;
    });
    out += `</ul>`;
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
            <td>${Object.values(checkValue(aef.securityMethods))}</td>
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


