export namespace drivers {
	
	export class ConnectionConfig {
	    Host: string;
	    Port: number;
	    User: string;
	    Password: string;
	    DriverType: string;
	    Database: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Host = source["Host"];
	        this.Port = source["Port"];
	        this.User = source["User"];
	        this.Password = source["Password"];
	        this.DriverType = source["DriverType"];
	        this.Database = source["Database"];
	    }
	}
	export class QueryResult {
	    Columns: string[];
	    Rows: any[][];
	
	    static createFrom(source: any = {}) {
	        return new QueryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Columns = source["Columns"];
	        this.Rows = source["Rows"];
	    }
	}

}

export namespace schema {
	
	export class Column {
	    Name: string;
	    Type: string;
	    Nullable: boolean;
	    Default: string;
	
	    static createFrom(source: any = {}) {
	        return new Column(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Type = source["Type"];
	        this.Nullable = source["Nullable"];
	        this.Default = source["Default"];
	    }
	}

}

