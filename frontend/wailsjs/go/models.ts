export namespace app {
	
	export class Dashboard {
	    points: store.SessionPoint[];
	    totals: store.Totals;
	    streak: number;
	    keys: stats.KeyStat[];
	    bigrams: stats.BigramStat[];
	
	    static createFrom(source: any = {}) {
	        return new Dashboard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.points = this.convertValues(source["points"], store.SessionPoint);
	        this.totals = this.convertValues(source["totals"], store.Totals);
	        this.streak = source["streak"];
	        this.keys = this.convertValues(source["keys"], stats.KeyStat);
	        this.bigrams = this.convertValues(source["bigrams"], stats.BigramStat);
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
	export class ExercisePlan {
	    segments: text.Segment[];
	    length: number;
	    position: number;
	    resume: text.Segment;
	    resumeStart: number;
	
	    static createFrom(source: any = {}) {
	        return new ExercisePlan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.segments = this.convertValues(source["segments"], text.Segment);
	        this.length = source["length"];
	        this.position = source["position"];
	        this.resume = this.convertValues(source["resume"], text.Segment);
	        this.resumeStart = source["resumeStart"];
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
	export class ExportResult {
	    path: string;
	    sessions: number;
	
	    static createFrom(source: any = {}) {
	        return new ExportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.sessions = source["sessions"];
	    }
	}
	export class PDFFile {
	    cancelled: boolean;
	    name: string;
	    data: number[];
	
	    static createFrom(source: any = {}) {
	        return new PDFFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cancelled = source["cancelled"];
	        this.name = source["name"];
	        this.data = source["data"];
	    }
	}
	export class ReviewExercise {
	    kind: string;
	    content: string;
	    words: string[];
	
	    static createFrom(source: any = {}) {
	        return new ReviewExercise(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.content = source["content"];
	        this.words = source["words"];
	    }
	}
	export class ReviewSummary {
	    due: number;
	    learning: number;
	    mastered: number;
	
	    static createFrom(source: any = {}) {
	        return new ReviewSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.due = source["due"];
	        this.learning = source["learning"];
	        this.mastered = source["mastered"];
	    }
	}
	export class SessionResult {
	    sessionId: number;
	    textId: number;
	    summary: stats.Summary;
	    mistakes: typing.WordError[];
	    average: store.Averages;
	
	    static createFrom(source: any = {}) {
	        return new SessionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.textId = source["textId"];
	        this.summary = this.convertValues(source["summary"], stats.Summary);
	        this.mistakes = this.convertValues(source["mistakes"], typing.WordError);
	        this.average = this.convertValues(source["average"], store.Averages);
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
	export class StartRequest {
	    textId: number;
	    start: number;
	    end: number;
	    content: string;
	    settings: string;
	
	    static createFrom(source: any = {}) {
	        return new StartRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.textId = source["textId"];
	        this.start = source["start"];
	        this.end = source["end"];
	        this.content = source["content"];
	        this.settings = source["settings"];
	    }
	}
	export class SystemInfo {
	    version: string;
	    platform: string;
	    databasePath: string;
	    schemaVersion: number;
	
	    static createFrom(source: any = {}) {
	        return new SystemInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.platform = source["platform"];
	        this.databasePath = source["databasePath"];
	        this.schemaVersion = source["schemaVersion"];
	    }
	}
	export class TextInput {
	    title: string;
	    content: string;
	    language: string;
	    tags: string[];
	    source: string;
	    sourceFile: string;
	    pages: string;
	
	    static createFrom(source: any = {}) {
	        return new TextInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.content = source["content"];
	        this.language = source["language"];
	        this.tags = source["tags"];
	        this.source = source["source"];
	        this.sourceFile = source["sourceFile"];
	        this.pages = source["pages"];
	    }
	}
	export class TextList {
	    texts: store.TextSummary[];
	    tags: string[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new TextList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.texts = this.convertValues(source["texts"], store.TextSummary);
	        this.tags = source["tags"];
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
	export class TextQuery {
	    search: string;
	    tag: string;
	    sort: string;
	
	    static createFrom(source: any = {}) {
	        return new TextQuery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.search = source["search"];
	        this.tag = source["tag"];
	        this.sort = source["sort"];
	    }
	}

}

export namespace srs {
	
	export class Card {
	    repetitions: number;
	    interval: number;
	    ease: number;
	    due: string;
	
	    static createFrom(source: any = {}) {
	        return new Card(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repetitions = source["repetitions"];
	        this.interval = source["interval"];
	        this.ease = source["ease"];
	        this.due = source["due"];
	    }
	}

}

export namespace stats {
	
	export class BigramStat {
	    bigram: string;
	    count: number;
	    avgMs: number;
	
	    static createFrom(source: any = {}) {
	        return new BigramStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bigram = source["bigram"];
	        this.count = source["count"];
	        this.avgMs = source["avgMs"];
	    }
	}
	export class KeyStat {
	    key: string;
	    count: number;
	    errors: number;
	    errorRate: number;
	    avgMs: number;
	
	    static createFrom(source: any = {}) {
	        return new KeyStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.count = source["count"];
	        this.errors = source["errors"];
	        this.errorRate = source["errorRate"];
	        this.avgMs = source["avgMs"];
	    }
	}
	export class Summary {
	    durationMs: number;
	    keystrokes: number;
	    correctKeystrokes: number;
	    backspaces: number;
	    correctChars: number;
	    uncorrectedErrors: number;
	    wpmGross: number;
	    wpmNet: number;
	    cpm: number;
	    accuracy: number;
	
	    static createFrom(source: any = {}) {
	        return new Summary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.durationMs = source["durationMs"];
	        this.keystrokes = source["keystrokes"];
	        this.correctKeystrokes = source["correctKeystrokes"];
	        this.backspaces = source["backspaces"];
	        this.correctChars = source["correctChars"];
	        this.uncorrectedErrors = source["uncorrectedErrors"];
	        this.wpmGross = source["wpmGross"];
	        this.wpmNet = source["wpmNet"];
	        this.cpm = source["cpm"];
	        this.accuracy = source["accuracy"];
	    }
	}

}

export namespace store {
	
	export class Averages {
	    sessions: number;
	    wpmNet: number;
	    accuracy: number;
	
	    static createFrom(source: any = {}) {
	        return new Averages(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessions = source["sessions"];
	        this.wpmNet = source["wpmNet"];
	        this.accuracy = source["accuracy"];
	    }
	}
	export class SessionPoint {
	    id: number;
	    endedAt: string;
	    mode: string;
	    textTitle: string;
	    durationMs: number;
	    wpmGross: number;
	    wpmNet: number;
	    cpm: number;
	    accuracy: number;
	    errors: number;
	
	    static createFrom(source: any = {}) {
	        return new SessionPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.endedAt = source["endedAt"];
	        this.mode = source["mode"];
	        this.textTitle = source["textTitle"];
	        this.durationMs = source["durationMs"];
	        this.wpmGross = source["wpmGross"];
	        this.wpmNet = source["wpmNet"];
	        this.cpm = source["cpm"];
	        this.accuracy = source["accuracy"];
	        this.errors = source["errors"];
	    }
	}
	export class Text {
	    id: number;
	    title: string;
	    content: string;
	    language: string;
	    source: string;
	    sourceFile: string;
	    pages: string;
	    tags: string[];
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Text(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.content = source["content"];
	        this.language = source["language"];
	        this.source = source["source"];
	        this.sourceFile = source["sourceFile"];
	        this.pages = source["pages"];
	        this.tags = source["tags"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class TextSummary {
	    id: number;
	    title: string;
	    language: string;
	    source: string;
	    sourceFile: string;
	    pages: string;
	    tags: string[];
	    length: number;
	    preview: string;
	    createdAt: string;
	    bestWpm?: number;
	    position: number;
	
	    static createFrom(source: any = {}) {
	        return new TextSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.language = source["language"];
	        this.source = source["source"];
	        this.sourceFile = source["sourceFile"];
	        this.pages = source["pages"];
	        this.tags = source["tags"];
	        this.length = source["length"];
	        this.preview = source["preview"];
	        this.createdAt = source["createdAt"];
	        this.bestWpm = source["bestWpm"];
	        this.position = source["position"];
	    }
	}
	export class Totals {
	    sessions: number;
	    totalMs: number;
	    bestWpm: number;
	
	    static createFrom(source: any = {}) {
	        return new Totals(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessions = source["sessions"];
	        this.totalMs = source["totalMs"];
	        this.bestWpm = source["bestWpm"];
	    }
	}
	export class Word {
	    id: number;
	    word: string;
	    shown: string;
	    encounters: number;
	    errors: number;
	    avgMsPerChar?: number;
	    mastery: number;
	    nextReview: string;
	    lastErrorAt: string;
	    card: srs.Card;
	    forms: string[];
	    dominantError: string;
	
	    static createFrom(source: any = {}) {
	        return new Word(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.word = source["word"];
	        this.shown = source["shown"];
	        this.encounters = source["encounters"];
	        this.errors = source["errors"];
	        this.avgMsPerChar = source["avgMsPerChar"];
	        this.mastery = source["mastery"];
	        this.nextReview = source["nextReview"];
	        this.lastErrorAt = source["lastErrorAt"];
	        this.card = this.convertValues(source["card"], srs.Card);
	        this.forms = source["forms"];
	        this.dominantError = source["dominantError"];
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

export namespace text {
	
	export class CleanOptions {
	    removePageNumbers: boolean;
	    removeRepeatedLines: boolean;
	    joinHyphenated: boolean;
	    normalizeApostrophes: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CleanOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.removePageNumbers = source["removePageNumbers"];
	        this.removeRepeatedLines = source["removeRepeatedLines"];
	        this.joinHyphenated = source["joinHyphenated"];
	        this.normalizeApostrophes = source["normalizeApostrophes"];
	    }
	}
	export class ImportPage {
	    number: number;
	    lines: string[];
	
	    static createFrom(source: any = {}) {
	        return new ImportPage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.number = source["number"];
	        this.lines = source["lines"];
	    }
	}
	export class Segment {
	    index: number;
	    start: number;
	    end: number;
	
	    static createFrom(source: any = {}) {
	        return new Segment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.start = source["start"];
	        this.end = source["end"];
	    }
	}

}

export namespace typing {
	
	export class Keystroke {
	    seq: number;
	    position: number;
	    typed: string;
	    timeMs: number;
	
	    static createFrom(source: any = {}) {
	        return new Keystroke(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.seq = source["seq"];
	        this.position = source["position"];
	        this.typed = source["typed"];
	        this.timeMs = source["timeMs"];
	    }
	}
	export class WordError {
	    expected: string;
	    typed: string;
	    position: number;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new WordError(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.expected = source["expected"];
	        this.typed = source["typed"];
	        this.position = source["position"];
	        this.type = source["type"];
	    }
	}

}

