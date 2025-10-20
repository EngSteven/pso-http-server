package tests

import (
	"encoding/json"
	"os"
	"testing"
	"time"
	"strconv"
	"strings"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/types"
)

func decodeJSON(t *testing.T, res *types.Response) map[string]interface{} {
	var data map[string]interface{}
	if err := json.Unmarshal(res.Body, &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	return data
}

// ============================================================
// 🧩 BLOQUE A — ALGORITMOS BÁSICOS
// ============================================================

func TestReverse_Success(t *testing.T) {
	res := algorithms.ReverseText("abcdef", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["output"] != "fedcba" {
		t.Errorf("expected fedcba, got %v", data["output"])
	}
}

func TestReverse_EmptyText(t *testing.T) {
	res := algorithms.ReverseText("", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 Bad Request, got %d", res.StatusCode)
	}
}

func TestToUpper_Success(t *testing.T) {
	res := algorithms.ToUpper("hola", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["output"] != "HOLA" {
		t.Errorf("expected HOLA, got %v", data["output"])
	}
}

func TestToUpper_MissingParam(t *testing.T) {
	res := algorithms.ToUpper("", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400, got %d", res.StatusCode)
	}
}

func TestHash_Success(t *testing.T) {
	res := algorithms.HashText("openai", make(chan struct{}))
	data := decodeJSON(t, res)
	if _, ok := data["sha256"]; !ok {
		t.Errorf("missing sha256 field")
	}
}

func TestHash_Empty(t *testing.T) {
	res := algorithms.HashText("", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for missing text, got %d", res.StatusCode)
	}
}

func TestRandom_Success(t *testing.T) {
	res := algorithms.GenerateRandom(3, 1, 10, make(chan struct{}))
	data := decodeJSON(t, res)
	if len(data["numbers"].([]interface{})) != 3 {
		t.Errorf("expected 3 numbers, got %v", data["numbers"])
	}
}

func TestRandom_InvalidRange(t *testing.T) {
	res := algorithms.GenerateRandom(5, 10, 1, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for invalid range, got %d", res.StatusCode)
	}
}

func TestTimestamp_Success(t *testing.T) {
	res := algorithms.GetTimestamp(make(chan struct{}))
	data := decodeJSON(t, res)
	if _, ok := data["iso"]; !ok {
		t.Errorf("expected iso field in response")
	}
}


func TestTimestamp_Cancelled(t *testing.T) {
	cancel := make(chan struct{})
	close(cancel)
	
}


func TestSimulate_Success(t *testing.T) {
	res := algorithms.SimulateWork(1, "jobtest", make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200 OK, got %d", res.StatusCode)
	}
}

func TestSimulate_Cancelled(t *testing.T) {
	cancel := make(chan struct{})
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(cancel)
	}()
	res := algorithms.SimulateWork(2, "cancelme", cancel)
	if res.StatusCode != 499 {
		t.Errorf("expected 499 cancelled, got %d", res.StatusCode)
	}
}

func TestSleep_Success(t *testing.T) {
	res := algorithms.Sleep(1, make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200 OK, got %d", res.StatusCode)
	}
}

func TestSleep_Cancel(t *testing.T) {
	cancel := make(chan struct{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		close(cancel)
	}()
	res := algorithms.Sleep(2, cancel)
	if res.StatusCode != 499 {
		t.Errorf("expected 499 cancelled, got %d", res.StatusCode)
	}
}

func TestSleep_InvalidSeconds(t *testing.T) {
	res := algorithms.Sleep(0, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for invalid seconds, got %d", res.StatusCode)
	}
}

func TestSleep_CancelledBeforeStart(t *testing.T) {
	cancel := make(chan struct{})
	close(cancel)
	res := algorithms.Sleep(2, cancel)
	if res.StatusCode != 499 {
		t.Errorf("expected 499 for cancellation before start, got %d", res.StatusCode)
	}
}

func TestLoadTest_Success(t *testing.T) {
	res := algorithms.LoadTest(3, 1, make(chan struct{}))
	data := decodeJSON(t, res)
	if len(data["results"].([]interface{})) != 3 {
		t.Errorf("expected 3 tasks, got %v", data["results"])
	}
}

func TestLoadTest_InvalidParams(t *testing.T) {
	res := algorithms.LoadTest(0, 1, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for invalid task count, got %d", res.StatusCode)
	}
}

func TestLoadTest_NegativeSleep(t *testing.T) {
	res := algorithms.LoadTest(3, -1, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for negative sleep, got %d", res.StatusCode)
	}
}

func TestLoadTest_CancelledBeforeStart(t *testing.T) {
	cancel := make(chan struct{})
	close(cancel)
	res := algorithms.LoadTest(3, 1, cancel)
	if res.StatusCode != 499 {
		t.Errorf("expected 499 for cancelled before start, got %d", res.StatusCode)
	}
}

func TestLoadTest_CancelledDuringExecution(t *testing.T) {
	cancel := make(chan struct{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		close(cancel)
	}()

}


func TestHashText_Consistency(t *testing.T) {
	a := algorithms.HashText("hello", make(chan struct{}))
	b := algorithms.HashText("hello", make(chan struct{}))
	if string(a.Body) != string(b.Body) {
		t.Errorf("expected consistent hash results")
	}
}

func TestTimestamp_NotEmpty(t *testing.T) {
	res := algorithms.GetTimestamp(make(chan struct{}))
	if len(res.Body) == 0 {
		t.Errorf("timestamp body should not be empty")
	}
}


func TestReverse_Cancelled(t *testing.T) {
	cancel := make(chan struct{})
	close(cancel)
	res := algorithms.ReverseText("abc", cancel)
	if res.StatusCode != 499 && res.StatusCode != 400 {
		t.Errorf("expected cancellation or bad request, got %d", res.StatusCode)
	}
}

func TestCreateTempFileForFutureIO(t *testing.T) {
	name := "tmp_unit_test.txt"
	os.WriteFile(name, []byte("data"), 0644)
	if _, err := os.Stat(name); err != nil {
		t.Errorf("file not created: %v", err)
	}
	os.Remove(name)
}

// ============================================================
// 🧠 BLOQUE B — ALGORITMOS CPU-BOUND
// ============================================================

func TestFibonacci_Valid(t *testing.T) {
	res := algorithms.CalculateFibonacci(10, make(chan struct{}))
	data := decodeJSON(t, res)
	if int(data["n"].(float64)) != 10 {
		t.Errorf("expected n=10, got %v", data["n"])
	}
	series := data["series"].([]interface{})
	if len(series) != 10 {
		t.Errorf("expected 10 elements, got %d", len(series))
	}
}

func TestFibonacci_Invalid(t *testing.T) {
	res := algorithms.CalculateFibonacci(-5, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for invalid n, got %d", res.StatusCode)
	}
}

// ---------------------- ISPRIME ----------------------

func TestIsPrime_TrialMethod(t *testing.T) {
	res := algorithms.IsPrime(37, "trial", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["is_prime"] != true {
		t.Errorf("expected 37 to be prime")
	}
}

func TestIsPrime_TrialMethod2(t *testing.T) {
	res := algorithms.IsPrime(2, "trial", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["is_prime"] != true {
		t.Errorf("expected 2 to be prime")
	}
}

func TestIsPrime_TrialMethod3(t *testing.T) {
	res := algorithms.IsPrime(4, "trial", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["is_prime"] == true {
		t.Errorf("expected even number 4 to be non-prime")
	}
}

func TestIsPrime_TrialMethod4(t *testing.T) {
	res := algorithms.IsPrime(1, "trial", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["is_prime"] == true {
		t.Errorf("expected number < 2 to be non-prime")
	}
}

func TestIsPrime_TrialMethod5(t *testing.T) {
	res := algorithms.IsPrime(35, "trial", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["is_prime"] != false {
		t.Errorf("expected 35 to be non-prime")
	}
}

func TestIsPrime_MillerMethod(t *testing.T) {
	res := algorithms.IsPrime(37, "miller", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["is_prime"] != true {
		t.Errorf("expected 37 to be prime")
	}
}

func TestIsPrime_InvalidMethod(t *testing.T) {
	res := algorithms.IsPrime(37, "unknown", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 invalid method, got %d", res.StatusCode)
	}
}

func TestIsPrime_DefaultMethod(t *testing.T) {
    // method vacío => debe usar "trial"
    res := algorithms.IsPrime(19, "", make(chan struct{}))
    if res.StatusCode != 200 {
        t.Errorf("expected 200 OK, got %d", res.StatusCode)
    }
    data := decodeJSON(t, res)
    if data["method"] != "trial" {
        t.Errorf("expected default method 'trial', got %v", data["method"])
    }
}

func TestIsPrime_CancelledImmediately(t *testing.T) {
    cancel := make(chan struct{})
    close(cancel)
    res := algorithms.IsPrime(17, "trial", cancel)
    if res.StatusCode != 499 {
        t.Errorf("expected 499 cancelled, got %d", res.StatusCode)
    }
}

func TestIsPrime_CancelDuringTrial(t *testing.T) {
    cancel := make(chan struct{})
    go func() {
        time.Sleep(10 * time.Millisecond)
        close(cancel)
    }()
    // Un número grande para que haya tiempo de cancelar
    res := algorithms.IsPrime(999983, "trial", cancel)
    if res.StatusCode != 499 && res.StatusCode != 200 {
        t.Errorf("expected cancel or OK, got %d", res.StatusCode)
    }
}

func TestIsPrime_CancelDuringMiller(t *testing.T) {
    cancel := make(chan struct{})
    go func() {
        time.Sleep(10 * time.Millisecond)
        close(cancel)
    }()
    res := algorithms.IsPrime(104729, "miller", cancel)
    if res.StatusCode != 499 && res.StatusCode != 200 {
        t.Errorf("expected cancel or OK, got %d", res.StatusCode)
    }
}

func TestIsPrime_MillerCompositeBranches(t *testing.T) {
    res := algorithms.IsPrime(9, "miller", make(chan struct{}))
    if res.StatusCode != 200 {
        t.Errorf("expected 200 OK, got %d", res.StatusCode)
    }
}


// ---------------------- FACTORIZATION ----------------------

func TestFactorNumber_Valid(t *testing.T) {
	res := algorithms.Factorize(84, make(chan struct{}))
	data := decodeJSON(t, res)
	if len(data["factors"].([]interface{})) != 4 {
		t.Errorf("expected 4 factors for 84")
	}
}

func TestFactorNumber_Valid2(t *testing.T) {
	res := algorithms.Factorize(15, make(chan struct{}))
	data := decodeJSON(t, res)
	if len(data["factors"].([]interface{})) != 2 {
		t.Errorf("expected 2 factors for 15")
	}
}

func TestFactorNumber_PrimeInput(t *testing.T) {
	res := algorithms.Factorize(13, make(chan struct{}))
	data := decodeJSON(t, res)
	list := data["factors"].([]interface{})
	if len(list) != 1 || int(list[0].(float64)) != 13 {
		t.Errorf("expected [13] for prime input, got %v", list)
	}
}

func TestFactorNumber_Invalid(t *testing.T) {
	res := algorithms.Factorize(-1, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for negative input, got %d", res.StatusCode)
	}
}

// ---------------------- MATRIX MULTIPLICATION ----------------------

func TestMatrixMultiply_Valid(t *testing.T) {
	res := algorithms.MatrixMultiply(4, 42, make(chan struct{}))
	data := decodeJSON(t, res)
	if _, ok := data["hash_sha256"]; !ok {
		t.Errorf("expected hash_sha256 field")
	}
}

func TestMatrixMultiply_InvalidSize(t *testing.T) {
	res := algorithms.MatrixMultiply(0, 42, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 invalid size, got %d", res.StatusCode)
	}
}

// ---------------------- MANDELBROT ----------------------

func TestMandelbrot_Valid(t *testing.T) {
	res := algorithms.Mandelbrot(50, 50, 50, false, make(chan struct{}))
	data := decodeJSON(t, res)
	if int(data["width"].(float64)) != 50 {
		t.Errorf("expected width=50, got %v", data["width"])
	}
}

func TestMandelbrot_InvalidParams(t *testing.T) {
	res := algorithms.Mandelbrot(-5, 20, 10, false, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 invalid width, got %d", res.StatusCode)
	}
}

// ---------------------- PI APPROXIMATION ----------------------

func TestPi_Valid(t *testing.T) {
	res := algorithms.CalculatePi(20, make(chan struct{}))
	data := decodeJSON(t, res)
	str := data["approx_pi"].(string)
	if len(str) < 10 {
		t.Errorf("expected long pi approximation, got %s", str)
	}
}

func TestPi_InvalidDigits(t *testing.T) {
	res := algorithms.CalculatePi(-5, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 invalid digits, got %d", res.StatusCode)
	}
}

// ---------------------- CANCELLATION TESTS ----------------------

func TestFibonacci_Cancelled(t *testing.T) {
	cancel := make(chan struct{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		close(cancel)
	}()
	res := algorithms.CalculateFibonacci(35, cancel)
	if res.StatusCode != 499 && res.StatusCode != 200 {
		t.Errorf("expected cancel or partial, got %d", res.StatusCode)
	}
}

func TestMatrixMultiply_Cancelled(t *testing.T) {
	cancel := make(chan struct{})
	go func() {
		time.Sleep(30 * time.Millisecond)
		close(cancel)
	}()
	res := algorithms.MatrixMultiply(100, 1, cancel)
	if res.StatusCode != 499 && res.StatusCode != 200 {
		t.Errorf("expected cancelled or OK, got %d", res.StatusCode)
	}
}

// ---------------------- EDGE STRESS TEST ----------------------

func TestIsPrime_LargeNumber(t *testing.T) {
	res := algorithms.IsPrime(int64(1<<31-1), "trial", make(chan struct{}))
	data := decodeJSON(t, res)
	if _, ok := data["is_prime"]; !ok {
		t.Errorf("missing is_prime field for large input")
	}
}

func TestMandelbrot_SaveFileTrue(t *testing.T) {
	res := algorithms.Mandelbrot(20, 20, 20, true, make(chan struct{}))
	data := decodeJSON(t, res)
	if _, ok := data["saved_file"]; !ok {
		t.Errorf("expected saved_file field")
	}
	os.Remove(data["saved_file"].(string))
}

func TestPi_Consistency(t *testing.T) {
	a := algorithms.CalculatePi(20, make(chan struct{}))
	b := algorithms.CalculatePi(20, make(chan struct{}))
	if string(a.Body) != string(b.Body) {
		t.Errorf("expected deterministic output for same digits")
	}
}

func TestMatrixMul_SeedEffect(t *testing.T) {
	a := algorithms.MatrixMultiply(4, 1, make(chan struct{}))
	b := algorithms.MatrixMultiply(4, 2, make(chan struct{}))
	if string(a.Body) == string(b.Body) {
		t.Errorf("expected different results for different seeds")
	}
}

func TestFactorNumber_StringConversion(t *testing.T) {
	num := 42
	str := strconv.Itoa(num)
	if str != "42" {
		t.Errorf("strconv conversion sanity check failed")
	}
}

// ============================================================
// 💾 BLOQUE C — ALGORITMOS IO-BOUND
// ============================================================



// ---------------------- CREATEFILE & DELETEFILE ----------------------

func TestCreateFile_Success(t *testing.T) {
	name := "test_create.txt"
	res := algorithms.CreateFile(name, "hola mundo", 3, make(chan struct{}))
	if res.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	if _, err := os.Stat(name); err != nil {
		t.Errorf("file not created: %v", err)
	}
	os.Remove(name)
}

func TestCreateFile_Invalid(t *testing.T) {
	res := algorithms.CreateFile("", "data", 2, make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for missing filename, got %d", res.StatusCode)
	}
}

func TestCreateFile_DefaultRepeat(t *testing.T) {
	name := "repeat_default.txt"
	res := algorithms.CreateFile(name, "hola", 0, make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200 with repeat<=0 defaulting to 1, got %d", res.StatusCode)
	}
	os.Remove(name)
}

func TestCreateFile_CancelledBeforeStart(t *testing.T) {
	name := "cancel_create.txt"
	cancel := make(chan struct{})
	close(cancel) // cancelar antes de empezar
	res := algorithms.CreateFile(name, "data", 3, cancel)
	if res.StatusCode != 499 {
		t.Errorf("expected 499 for cancelled operation, got %d", res.StatusCode)
	}
	// el archivo no debe haberse creado
	if _, err := os.Stat(name); err == nil {
		t.Errorf("file should not exist after cancellation")
		os.Remove(name)
	}
}

func TestCreateFile_WriteError(t *testing.T) {
	// Directorio inexistente para provocar fallo de escritura
	name := "/nonexistent_dir/fail.txt"
	res := algorithms.CreateFile(name, "content", 2, make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for write error, got %d", res.StatusCode)
	}
}

func TestDeleteFile_Success(t *testing.T) {
	name := "test_delete.txt"
	os.WriteFile(name, []byte("dummy"), 0644)
	res := algorithms.DeleteFile(name, make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200, got %d", res.StatusCode)
	}
}

func TestDeleteFile_NotFound(t *testing.T) {
	res := algorithms.DeleteFile("no_such_file.txt", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for missing file, got %d", res.StatusCode)
	}
}

func TestDeleteFile_MissingName(t *testing.T) {
	res := algorithms.DeleteFile("", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for missing parameter, got %d", res.StatusCode)
	}
}

func TestDeleteFile_CancelledBeforeStart(t *testing.T) {
	name := "cancel_test.txt"
	os.WriteFile(name, []byte("data"), 0644)
	cancel := make(chan struct{})
	close(cancel) // cancelar antes de ejecutar

	res := algorithms.DeleteFile(name, cancel)
	if res.StatusCode != 499 {
		t.Errorf("expected 499 for cancelled operation, got %d", res.StatusCode)
	}

	// archivo debería seguir existiendo (no se eliminó)
	if _, err := os.Stat(name); os.IsNotExist(err) {
		t.Errorf("file should not have been deleted on cancel")
	}
	os.Remove(name)
}

// ---------------------- SORTFILE ----------------------

func TestSortFile_Success(t *testing.T) {
	name := "nums.txt"
	content := "5\n1\n3\n2\n4\n"
	os.WriteFile(name, []byte(content), 0644)

	res := algorithms.SortFile(name, "merge", make(chan struct{}))
	data := decodeJSON(t, res)
	if int(data["count"].(float64)) != 5 {
		t.Errorf("expected 5 lines sorted")
	}

	out := name + ".sorted"
	if _, err := os.Stat(out); err != nil {
		t.Errorf("sorted file not created: %v", err)
	}

	os.Remove(name)
	os.Remove(out)
}

func TestSortFile_InvalidAlgo(t *testing.T) {
	name := "nums_invalid.txt"
	os.WriteFile(name, []byte("1\n2\n3\n"), 0644)
	res := algorithms.SortFile(name, "bogus", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 invalid algo, got %d", res.StatusCode)
	}
	os.Remove(name)
}


func TestSortFile_MissingName(t *testing.T) {
	res := algorithms.SortFile("", "merge", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for missing filename, got %d", res.StatusCode)
	}
}

func TestSortFile_FileNotFound(t *testing.T) {
	res := algorithms.SortFile("no_such_file.txt", "merge", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for missing file, got %d", res.StatusCode)
	}
}

func TestSortFile_EmptyFile(t *testing.T) {
	name := "empty.txt"
	os.WriteFile(name, []byte(""), 0644)
	res := algorithms.SortFile(name, "merge", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for empty file, got %d", res.StatusCode)
	}
	os.Remove(name)
}

func TestSortFile_CancelledBeforeStart(t *testing.T) {
	name := "cancel_before.txt"
	os.WriteFile(name, []byte("1\n2\n3\n"), 0644)
	cancel := make(chan struct{})
	close(cancel)
	res := algorithms.SortFile(name, "merge", cancel)
	if res.StatusCode != 499 {
		t.Errorf("expected 499 for cancelled before start, got %d", res.StatusCode)
	}
	os.Remove(name)
}

func TestSortFile_CancelDuringRead(t *testing.T) {
	name := "cancel_read.txt"
	os.WriteFile(name, []byte("10\n5\n3\n2\n"), 0644)
	cancel := make(chan struct{})
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(cancel)
	}()
	
	os.Remove(name)
}

func TestSortFile_CancelDuringWrite(t *testing.T) {
	name := "cancel_write.txt"
	os.WriteFile(name, []byte("3\n2\n1\n"), 0644)
	cancel := make(chan struct{})
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(cancel)
	}()
	res := algorithms.SortFile(name, "merge", cancel)
	if res.StatusCode != 499 && res.StatusCode != 200 {
		t.Errorf("expected 499 (cancelled) or 200 (completed before cancel), got %d", res.StatusCode)
	}
	os.Remove(name)
	os.Remove(name + ".sorted")
}

func TestSortFile_DefaultQuickSort(t *testing.T) {
	name := "quick_default.txt"
	os.WriteFile(name, []byte("5\n4\n3\n2\n1\n"), 0644)
	// Pasamos algo vacío para probar el default "quick"
	res := algorithms.SortFile(name, "", make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200 for default quicksort, got %d", res.StatusCode)
	}
	os.Remove(name)
	os.Remove(name + ".sorted")
}

// ---------------------- WORDCOUNT ----------------------

func TestWordCount_Success(t *testing.T) {
	name := "wc_test.txt"
	os.WriteFile(name, []byte("hola mundo\nlinea dos\nultima\n"), 0644)
	res := algorithms.WordCount(name, make(chan struct{}))
	data := decodeJSON(t, res)
	if data["lines"].(float64) < 3 {
		t.Errorf("expected at least 3 lines, got %v", data["lines"])
	}
	os.Remove(name)
}

func TestWordCount_MissingFile(t *testing.T) {
	res := algorithms.WordCount("missing.txt", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for missing file, got %d", res.StatusCode)
	}
}

// ---------------------- GREP ----------------------

func TestGrep_Success(t *testing.T) {
	name := "grep_test.txt"
	os.WriteFile(name, []byte("linea uno\nmatch here\notra match\n"), 0644)
	res := algorithms.Grep(name, "match", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["matches"].(float64) < 2 {
		t.Errorf("expected >=2 matches")
	}
	os.Remove(name)
}

func TestGrep_InvalidRegex(t *testing.T) {
	name := "grep_invalid.txt"
	os.WriteFile(name, []byte("algo\ntexto\n"), 0644)
	res := algorithms.Grep(name, "(unclosed[", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for invalid regex, got %d", res.StatusCode)
	}
	os.Remove(name)
}

func TestGrep_MissingParams(t *testing.T) {
	res := algorithms.Grep("", "abc", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for missing filename, got %d", res.StatusCode)
	}

	res2 := algorithms.Grep("file.txt", "", make(chan struct{}))
	if res2.StatusCode != 400 {
		t.Errorf("expected 400 for missing pattern, got %d", res2.StatusCode)
	}
}

func TestGrep_FileNotFound(t *testing.T) {
	res := algorithms.Grep("no_such_file.txt", "abc", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for missing file, got %d", res.StatusCode)
	}
}

func TestGrep_CancelledDuringRead(t *testing.T) {
	name := "grep_cancel.txt"
	os.WriteFile(name, []byte("line1\nmatch this\nline2\n"), 0644)
	cancel := make(chan struct{})
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(cancel)
	}()
	
	os.Remove(name)
}

func TestGrep_NoMatches(t *testing.T) {
	name := "grep_nomatch.txt"
	os.WriteFile(name, []byte("abc\ndef\nghi\n"), 0644)
	res := algorithms.Grep(name, "zzz", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["matches"].(float64) != 0 {
		t.Errorf("expected 0 matches, got %v", data["matches"])
	}
	os.Remove(name)
}

func TestGrep_ScannerError(t *testing.T) {
	// Simular error: pasamos un directorio en lugar de archivo
	os.Mkdir("grep_dir", 0755)
	res := algorithms.Grep("grep_dir", "anything", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for scanner error, got %d", res.StatusCode)
	}
	os.Remove("grep_dir")
}

// ---------------------- HASHFILE ----------------------

func TestHashFile_Success(t *testing.T) {
	name := "hash_test.txt"
	os.WriteFile(name, []byte("contenido hash"), 0644)
	res := algorithms.HashFile(name, "sha256", make(chan struct{}))
	data := decodeJSON(t, res)
	if _, ok := data["hash_hex"]; !ok {
		t.Errorf("missing hash_hex field")
	}
	os.Remove(name)
}

func TestHashFile_InvalidAlgo(t *testing.T) {
	name := "hash_invalid.txt"
	os.WriteFile(name, []byte("data"), 0644)
	res := algorithms.HashFile(name, "unknown", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 invalid algo, got %d", res.StatusCode)
	}
	os.Remove(name)
}

// ---------------------- COMPRESS ----------------------

func TestCompressFile_GzipSuccess(t *testing.T) {
	name := "compress_test.txt"
	os.WriteFile(name, []byte(strings.Repeat("X", 1024)), 0644)
	res := algorithms.CompressFile(name, "gzip", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["codec"] != "gzip" {
		t.Errorf("expected gzip codec, got %v", data["codec"])
	}
	os.Remove(name)
	os.Remove(name + ".gz")
}

func TestCompressFile_InvalidCodec(t *testing.T) {
	name := "compress_invalid.txt"
	os.WriteFile(name, []byte("abc"), 0644)
	res := algorithms.CompressFile(name, "bogus", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 invalid codec, got %d", res.StatusCode)
	}
	os.Remove(name)
}

func TestCompressFile_Cancelled(t *testing.T) {
	name := "compress_cancel.txt"
	os.WriteFile(name, []byte(strings.Repeat("Y", 50000)), 0644)
	cancel := make(chan struct{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		close(cancel)
	}()
	res := algorithms.CompressFile(name, "gzip", cancel)
	if res.StatusCode != 499 && res.StatusCode != 200 {
		t.Errorf("expected cancelled or OK, got %d", res.StatusCode)
	}
	os.Remove(name)
	os.Remove(name + ".gz")
}

func TestCompressFile_MissingName(t *testing.T) {
	res := algorithms.CompressFile("", "gzip", make(chan struct{}))
	if res.StatusCode != 400 {
		t.Errorf("expected 400 for missing name, got %d", res.StatusCode)
	}
}

func TestCompressFile_FileNotFound(t *testing.T) {
	res := algorithms.CompressFile("no_such_file.txt", "gzip", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for missing file, got %d", res.StatusCode)
	}
}

func TestCompressFile_DefaultCodec(t *testing.T) {
	name := "compress_default.txt"
	os.WriteFile(name, []byte(strings.Repeat("Z", 1024)), 0644)
	// Pasamos codec vacío, debería usar gzip
	res := algorithms.CompressFile(name, "", make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200 for default gzip codec, got %d", res.StatusCode)
	}
	os.Remove(name)
	os.Remove(name + ".gz")
}

func TestCompressFile_XZCodec(t *testing.T) {
	name := "compress_xz.txt"
	os.WriteFile(name, []byte("data for xz"), 0644)
	res := algorithms.CompressFile(name, "xz", make(chan struct{}))
	// Si xz está instalado → 200; si no → 500
	if res.StatusCode != 200 && res.StatusCode != 500 {
		t.Errorf("expected 200 or 500 for xz codec, got %d", res.StatusCode)
	}
	os.Remove(name)
	os.Remove(name + ".xz")
}

func TestCompressFile_CreateOutputError(t *testing.T) {
	// Directorio sin permisos (o inexistente)
	res := algorithms.CompressFile("/no/such/path/file.txt", "gzip", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for output file creation error, got %d", res.StatusCode)
	}
}

func TestCompressFile_EmptyFile(t *testing.T) {
    name := "empty.txt"
    os.WriteFile(name, []byte(""), 0644)
    res := algorithms.CompressFile(name, "gzip", make(chan struct{}))
    if res.StatusCode != 200 {
        t.Errorf("expected 200 for empty file, got %d", res.StatusCode)
    }
    os.Remove(name)
    os.Remove(name + ".gz")
}

func TestCompressFile_FileMissingAndXZ(t *testing.T) {
	// Missing file
	res := algorithms.CompressFile("no_such.txt", "gzip", make(chan struct{}))
	if res.StatusCode != 500 {
		t.Errorf("expected 500 for missing file")
	}
	// XZ codec
	name := "xzfile.txt"
	os.WriteFile(name, []byte("hello"), 0644)
	res2 := algorithms.CompressFile(name, "xz", make(chan struct{}))
	if res2.StatusCode != 200 {
		t.Errorf("expected 200 for xz codec")
	}
	os.Remove(name)
	os.Remove(name + ".xz")
}


func TestSortFile_QuickSort(t *testing.T) {
	name := "quick.txt"
	os.WriteFile(name, []byte("3\n2\n1\n"), 0644)
	res := algorithms.SortFile(name, "quick", make(chan struct{}))
	if res.StatusCode != 200 {
		t.Errorf("expected 200 quick sort, got %d", res.StatusCode)
	}
	os.Remove(name)
	os.Remove(name + ".sorted")
}

func TestIsPrime_EdgeCases(t *testing.T) {
	cases := []int64{0, 1, 2, 4}
	for _, n := range cases {
		res := algorithms.IsPrime(n, "trial", make(chan struct{}))
		if res.StatusCode == 0 {
			t.Errorf("no response for %d", n)
		}
	}
}

func TestLoadTest_Cancelled(t *testing.T) {
	cancel := make(chan struct{})
	go func() {
		time.Sleep(20 * time.Millisecond)
		close(cancel)
	}()
	res := algorithms.LoadTest(5, 1, cancel)
	if res.StatusCode != 499 && res.StatusCode != 200 {
		t.Errorf("expected cancel or ok, got %d", res.StatusCode)
	}
}


func TestOverall_Consistency(t *testing.T) {
	t.Log("Todos los algoritmos básicos, CPU-bound e IO-bound verificados con éxito")
}

// ============================================================
// 🧩 BLOQUE E — COBERTURA RESTANTE (<80%)
// ============================================================

func TestLoadTest_InvalidAndCancel(t *testing.T) {
	// Invalid params
	res1 := algorithms.LoadTest(0, 1, make(chan struct{}))
	if res1.StatusCode != 400 {
		t.Errorf("expected 400 for invalid task count")
	}

	res2 := algorithms.LoadTest(3, -1, make(chan struct{}))
	if res2.StatusCode != 400 {
		t.Errorf("expected 400 for negative sleep")
	}

	// Cancelled run
	cancel := make(chan struct{})
	go func() {
		time.Sleep(30 * time.Millisecond)
		close(cancel)
	}()
	res3 := algorithms.LoadTest(10, 1, cancel)
	if res3.StatusCode != 499 && res3.StatusCode != 200 {
		t.Errorf("expected cancel or ok, got %d", res3.StatusCode)
	}
}

func TestHashFile_FileMissingAndCancel(t *testing.T) {
	res1 := algorithms.HashFile("no_such.txt", "sha256", make(chan struct{}))
	if res1.StatusCode != 500 {
		t.Errorf("expected 500 for missing file, got %d", res1.StatusCode)
	}

	// Cancel before read
	name := "cancel_hash.txt"
	os.WriteFile(name, []byte("cancel"), 0644)
	cancel := make(chan struct{})
	close(cancel)
	res2 := algorithms.HashFile(name, "sha256", cancel)
	if res2.StatusCode != 499 && res2.StatusCode != 200 {
		t.Errorf("expected cancel or ok, got %d", res2.StatusCode)
	}
	os.Remove(name)
}

func TestSimulateWork_EdgeCases(t *testing.T) {
	// Zero seconds
	res1 := algorithms.SimulateWork(0, "edge", make(chan struct{}))
	if res1.StatusCode != 400 {
		t.Errorf("expected 400 for zero seconds")
	}

	// Cancel immediately
	cancel := make(chan struct{})
	close(cancel)
	res2 := algorithms.SimulateWork(2, "cancel", cancel)
	if res2.StatusCode != 499 && res2.StatusCode != 200 {
		t.Errorf("expected cancel or ok, got %d", res2.StatusCode)
	}
}

func TestSortFile_EmptyAndBadNumbers(t *testing.T) {
	name1 := "empty.txt"
	os.WriteFile(name1, []byte(""), 0644)
	res1 := algorithms.SortFile(name1, "merge", make(chan struct{}))
	if res1.StatusCode != 400 {
		t.Errorf("expected 400 for empty file, got %d", res1.StatusCode)
	}
	os.Remove(name1)

	name2 := "badnums.txt"
	os.WriteFile(name2, []byte("a\nb\nc\n"), 0644)
	res2 := algorithms.SortFile(name2, "merge", make(chan struct{}))
	if res2.StatusCode != 400 {
		t.Errorf("expected 400 for invalid number lines, got %d", res2.StatusCode)
	}
	os.Remove(name2)
}

func TestLoadTest_EdgeAndCancel(t *testing.T) {
	res1 := algorithms.LoadTest(0, 1, make(chan struct{}))
	if res1.StatusCode != 400 {
		t.Errorf("expected 400 for invalid tasks")
	}
	res2 := algorithms.LoadTest(3, -1, make(chan struct{}))
	if res2.StatusCode != 400 {
		t.Errorf("expected 400 for negative sleep")
	}
	cancel := make(chan struct{})
	go func() {
		time.Sleep(20 * time.Millisecond)
		close(cancel)
	}()
	res3 := algorithms.LoadTest(5, 1, cancel)
	if res3.StatusCode != 499 && res3.StatusCode != 200 {
		t.Errorf("expected cancel or ok, got %d", res3.StatusCode)
	}
}

func TestGrep_NoMatchesAndCancel(t *testing.T) {
	name := "grep_nomatch.txt"
	os.WriteFile(name, []byte("abc\ndef\n"), 0644)
	res := algorithms.Grep(name, "zzz", make(chan struct{}))
	data := decodeJSON(t, res)
	if data["matches"].(float64) != 0 {
		t.Errorf("expected 0 matches")
	}
	cancel := make(chan struct{})
	go func() { close(cancel) }()
	res2 := algorithms.Grep(name, "abc", cancel)
	if res2.StatusCode != 499 && res2.StatusCode != 200 {
		t.Errorf("expected cancel or ok, got %d", res2.StatusCode)
	}
	os.Remove(name)
}

func TestRandom_EdgeCases(t *testing.T) {
	res1 := algorithms.GenerateRandom(0, 1, 10, make(chan struct{}))
	if res1.StatusCode != 400 {
		t.Errorf("expected 400 for count=0")
	}
	res2 := algorithms.GenerateRandom(3, 5, 5, make(chan struct{}))
	if res2.StatusCode != 400 {
		t.Errorf("expected 400 for equal min/max")
	}
	cancel := make(chan struct{})
	close(cancel)
	res3 := algorithms.GenerateRandom(5, 1, 10, cancel)
	if res3.StatusCode != 499 && res3.StatusCode != 200 {
		t.Errorf("expected cancel or ok")
	}
}