import { keycloak } from "./keycloak";

class Api {
    async getPageData() {
        await keycloak.updateToken(30);
        const resp = await fetch('http://localhost:11488/api/v1/getPageData', {
            headers: {
              'Authorization': `Bearer: ${keycloak.token}`,
              'Accept': 'application/json'
            },
        });
        return await resp.json();
    }
}

export default new Api();