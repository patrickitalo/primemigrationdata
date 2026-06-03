export namespace app {
	
	export class EnvDefaults {
	    firebird: config.FirebirdFormEnv;
	    pharmacie: config.PharmacieFormEnv;
	
	    static createFrom(source: any = {}) {
	        return new EnvDefaults(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.firebird = this.convertValues(source["firebird"], config.FirebirdFormEnv);
	        this.pharmacie = this.convertValues(source["pharmacie"], config.PharmacieFormEnv);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class HistoryStatus {
	    connected: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new HistoryStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.error = source["error"];
	    }
	}
	export class LastRunInfo {
	    runId: string;
	    options: string;
	    mode: string;
	    status: string;
	    startedAt: string;
	    finishedAt?: string;
	    implantador: string;
	
	    static createFrom(source: any = {}) {
	        return new LastRunInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runId = source["runId"];
	        this.options = source["options"];
	        this.mode = source["mode"];
	        this.status = source["status"];
	        this.startedAt = source["startedAt"];
	        this.finishedAt = source["finishedAt"];
	        this.implantador = source["implantador"];
	    }
	}
	export class LoginResult {
	    session?: models.UserSession;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new LoginResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.session = this.convertValues(source["session"], models.UserSession);
	        this.error = source["error"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class OptionDef {
	    code: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new OptionDef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	    }
	}

}

export namespace config {
	
	export class FirebirdFormEnv {
	    Host: string;
	    Port: string;
	    Path: string;
	    User: string;
	    Password: string;
	    Conversao: string;
	
	    static createFrom(source: any = {}) {
	        return new FirebirdFormEnv(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Host = source["Host"];
	        this.Port = source["Port"];
	        this.Path = source["Path"];
	        this.User = source["User"];
	        this.Password = source["Password"];
	        this.Conversao = source["Conversao"];
	    }
	}
	export class PharmacieFormEnv {
	    Alias: string;
	    IPServer: string;
	    Porta: string;
	
	    static createFrom(source: any = {}) {
	        return new PharmacieFormEnv(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Alias = source["Alias"];
	        this.IPServer = source["IPServer"];
	        this.Porta = source["Porta"];
	    }
	}

}

export namespace models {
	
	export class ClientConfig {
	    id: number;
	    codigo_cliente: string;
	    sistema_origem: string;
	    db_path: string;
	    db_user: string;
	    db_password: string;
	    db_host: string;
	    db_port: string;
	    alias_pharmacie: string;
	    ipserver_pharmacie: string;
	    porta_pharmacie: string;
	    conversao: string;
	
	    static createFrom(source: any = {}) {
	        return new ClientConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.codigo_cliente = source["codigo_cliente"];
	        this.sistema_origem = source["sistema_origem"];
	        this.db_path = source["db_path"];
	        this.db_user = source["db_user"];
	        this.db_password = source["db_password"];
	        this.db_host = source["db_host"];
	        this.db_port = source["db_port"];
	        this.alias_pharmacie = source["alias_pharmacie"];
	        this.ipserver_pharmacie = source["ipserver_pharmacie"];
	        this.porta_pharmacie = source["porta_pharmacie"];
	        this.conversao = source["conversao"];
	    }
	}
	export class DatabaseConfig {
	    host: string;
	    port: string;
	    path: string;
	    user: string;
	    password: string;
	    conversao: string;
	
	    static createFrom(source: any = {}) {
	        return new DatabaseConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.port = source["port"];
	        this.path = source["path"];
	        this.user = source["user"];
	        this.password = source["password"];
	        this.conversao = source["conversao"];
	    }
	}
	export class MigrationConfig {
	    client_code: string;
	    system: string;
	    database: DatabaseConfig;
	    options: string[];
	    mode: string;
	    v_vencido?: string;
	    excel_path: string;
	    alias_pharmacie: string;
	    ipserver_pharmacie: string;
	    porta_pharmacie: string;
	
	    static createFrom(source: any = {}) {
	        return new MigrationConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.client_code = source["client_code"];
	        this.system = source["system"];
	        this.database = this.convertValues(source["database"], DatabaseConfig);
	        this.options = source["options"];
	        this.mode = source["mode"];
	        this.v_vencido = source["v_vencido"];
	        this.excel_path = source["excel_path"];
	        this.alias_pharmacie = source["alias_pharmacie"];
	        this.ipserver_pharmacie = source["ipserver_pharmacie"];
	        this.porta_pharmacie = source["porta_pharmacie"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class UserSession {
	    UserID: number;
	    Nome: string;
	    Nick: string;
	    Departamento: string;
	    Funcao: string;
	
	    static createFrom(source: any = {}) {
	        return new UserSession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.UserID = source["UserID"];
	        this.Nome = source["Nome"];
	        this.Nick = source["Nick"];
	        this.Departamento = source["Departamento"];
	        this.Funcao = source["Funcao"];
	    }
	}

}

