import { fuzzySearch } from "./fuzzySearch";

const tests: Record<string, Function | {run: Function}> = {
    emptyTestString() {
        let search =  "the quick brown fox jumps over the lazy"
        let result = fuzzySearch("", search);
        check();
        search = "0123456789";
        result = fuzzySearch("", search);
        check();
        result = fuzzySearch("", search);
        check();
        result = fuzzySearch("", search);
        check();

        function check() {
            const expectedScore = search === "" ? 1 : 0;
            if (
                result.index === 0
                && result.length === 0
                && result.minimumEditDistance === search.length
                && result.score === expectedScore
            ) {
                // test passed
            } else {
                throw {result, expectedScore, search};
            }
        }
    },
    emptySearchString() {
        let result = fuzzySearch("the quick brown fox jumps over the lazy dog", "");
        check();
        result = fuzzySearch("0123456789", "");
        check();
        result = fuzzySearch("0", "");
        check();
        result = fuzzySearch("", "");
        check();

        function check() {
            if (!(
                result.index === 0
                && result.length === 0
                && result.minimumEditDistance === 0
                && result.score === 1
            )){
                throw result;
            }
        }
    },
    searchStringGEQTestString() {
        let params: Parameters<typeof fuzzySearch> = ["", "abc"];
        let result = fuzzySearch(...params);
        let expectedResult:  ReturnType<(typeof fuzzySearch) & {}> = {
            index: 0,
            length: 0,
            minimumEditDistance: 3,
            score: 0,
        }
        check();
        params = ["fox", "foxes"]
        result = fuzzySearch(...params);
        expectedResult = {
            index: 0,
            length: 3,
            minimumEditDistance: 2,
            score: 3 / 5
        }
        check();
        params = ["fax", "foxes"]
        result = fuzzySearch(...params);
        expectedResult = {
            index: 0,
            length: 3,
            minimumEditDistance: 3,
            score: 2 / 5
        }
        check();
        params = ["fax", "fox"]
        result = fuzzySearch(...params);
        expectedResult = {
            index: 0,
            length: 3,
            minimumEditDistance: 1,
            score: 2 / 3
        }
        check();

        function check() {
            for (const field in expectedResult) {
                if ((expectedResult as any)[field] !== (result as any)[field]){
                    throw {params, result, expectedResult}
                }
            }
        }
    }
};


for (const testName in tests) {
    const test = tests[testName];
    try {
        if (test instanceof Function) {
            test();
        } else {
            test.run();
        }
        console.log("PASSED: " + testName);
    }
    catch(error) {
        console.error("FAILED: " + testName, error, test);
    }
}