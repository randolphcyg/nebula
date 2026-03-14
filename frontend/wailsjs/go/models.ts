export namespace gorm {
	
	export class DeletedAt {
	    // Go type: time
	    Time: any;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DeletedAt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Time = this.convertValues(source["Time"], null);
	        this.Valid = source["Valid"];
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

}

export namespace models {
	
	export class PcapFile {
	    id: number;
	    fileId: string;
	    fileName: string;
	    filePath: string;
	    fileSize: string;
	    status: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new PcapFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.fileId = source["fileId"];
	        this.fileName = source["fileName"];
	        this.filePath = source["filePath"];
	        this.fileSize = source["fileSize"];
	        this.status = source["status"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
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

}

export namespace pcap {
	
	export class FileQueryReq {
	    fileName: string;
	    fileSize: string;
	    startDate: string;
	    endDate: string;
	    page: number;
	    pageSize: number;
	
	    static createFrom(source: any = {}) {
	        return new FileQueryReq(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fileName = source["fileName"];
	        this.fileSize = source["fileSize"];
	        this.startDate = source["startDate"];
	        this.endDate = source["endDate"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	    }
	}
	export class FileQueryResp {
	    list: models.PcapFile[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new FileQueryResp(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.list = this.convertValues(source["list"], models.PcapFile);
	        this.total = source["total"];
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

}

export namespace zeek {
	
	export class AnalyzePCAPRequest {
	    TaskID: string;
	    UUID: string;
	    OnlyNotice: boolean;
	    PcapID: string;
	    PcapPath: string;
	    ScriptID: string;
	    ScriptPath: string;
	    ExtractedFilePath: string;
	    ExtractedFileMinSize: number;
	
	    static createFrom(source: any = {}) {
	        return new AnalyzePCAPRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TaskID = source["TaskID"];
	        this.UUID = source["UUID"];
	        this.OnlyNotice = source["OnlyNotice"];
	        this.PcapID = source["PcapID"];
	        this.PcapPath = source["PcapPath"];
	        this.ScriptID = source["ScriptID"];
	        this.ScriptPath = source["ScriptPath"];
	        this.ExtractedFilePath = source["ExtractedFilePath"];
	        this.ExtractedFileMinSize = source["ExtractedFileMinSize"];
	    }
	}
	export class AnalyzeResult {
	    TaskID: string;
	    Status: string;
	    StartTime: string;
	
	    static createFrom(source: any = {}) {
	        return new AnalyzeResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TaskID = source["TaskID"];
	        this.Status = source["Status"];
	        this.StartTime = source["StartTime"];
	    }
	}

}

