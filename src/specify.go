package specify

import(
	"fmt";
	"container/list";
	"os";
)

type Specification interface {
	Run();
	Describe(name string, description func());
	It(name string, description func());
	That(value Value) *That;
}

type specification struct {
	currentDescribe *describe;
	currentIt *it;
	describes *list.List;
}

func (self *specification) Run() {
	report := makeReport();
	runList(self.describes, report);
	fmt.Println(report.summary());
	if report.failed > 0 {
		os.Exit(1);
	}
}

func (self *specification) Describe(name string, description func()) {
	self.currentDescribe = makeDescribe(name);
	description();
	self.describes.PushBack(self.currentDescribe);
}

func (self *specification) It(name string, description func()) {
	self.currentIt = makeIt(name);
	description();
	self.currentDescribe.addIt(self.currentIt);
}

type Value interface{}

func (self *specification) That(value Value) (result *That) {
	result = makeThat(self.currentDescribe, self.currentIt, value);
	self.currentIt.addThat(result);
	return;
}

func New() Specification {
	return &specification{describes:list.New()};
}
 
type describe struct {
	name string;
	its *list.List;
}

func makeDescribe(name string) (result *describe) {
	result = &describe{name:name};
	result.its = list.New();
	return;
}

func (self *describe) addIt(it *it) {
	self.its.PushBack(it);
}

func (self *describe) run(report *report) {
	runList(self.its, report);
}

type it struct {
	name string;
	thats *list.List;
}

func makeIt(name string) (result *it) {
	result = &it{name:name};
	result.thats = list.New();
	return;
}

func (self *it) addThat(that *That) {
	self.thats.PushBack(that);
}

func (self *it) run(report *report) {
	runList(self.thats, report);
}

type That struct {
	Should Matcher;
	ShouldNot Matcher;
}

func makeThat(describe *describe, it *it, value Value) *That {
	matcher := &matcher{describe:describe, it:it, value:value};
	return &That{
		&should{matcher},
		&shouldNot{matcher}
	};
}

func (self *That) run(report *report) {
	runner,_ := self.Should.(runner);
	runner.run(report);
}

type Matcher interface {
	Be(Value);
}

type matcher struct {
	*describe;
	*it;
	value Value;
	block func() (bool, string);
}

func (self *matcher) run(report *report) {
	if pass,msg := self.block(); pass {
		report.pass();
	} else {
		report.fail(msg);
	}
}

type should struct { *matcher }
func (self *should) Be(value Value) {
	self.matcher.block = func() (bool, string) {
		if self.matcher.value != value {
			error := fmt.Sprintf("%v - %v - expected `%v` to be `%v`", self.matcher.describe.name, self.matcher.it.name, self.matcher.value, value);
			return false, error;
		}
		return true, "";
	}
}

type shouldNot struct { *matcher }
func (self *shouldNot) Be(value Value) {
	self.matcher.block = func() (bool, string) {
		if self.matcher.value == value {
			error := fmt.Sprintf("%v - %v - expected `%v` not to be `%v`", self.matcher.describe.name, self.matcher.it.name, self.matcher.value, value);
			return false, error;
		}
		return true, "";
	}
}

type runner interface {
	run(*report);
}

func runList(list *list.List, report *report) {
	doList(list, func(item Value) {
		runner,_ := item.(runner);
		runner.run(report);
	});
}

func doList(list *list.List, do func(Value)) {
	iter := list.Iter();
	for !closed(iter) {
		item := <-iter;
		if item == nil { break; }
		do(item);
	}
}

type report struct {
	passed, failed int;
	failures *list.List;
}

func makeReport() *report {
	return &report{failures:list.New()};
}

func (self *report) pass() { self.passed++; }
func (self *report) total() int { return self.passed + self.failed; }

func (self *report) fail(msg string) {
	self.failures.PushBack(msg);
	self.failed++;
}

func (self *report) summary() string {
	if self.failed > 0 {
		fmt.Println("\nFAILED TESTS:");
		doList(self.failures, func(item Value) { fmt.Println("-", item) });
		fmt.Println("");
	}
	return fmt.Sprintf("Passed: %v Failed: %v Total: %v", self.passed, self.failed, self.total());
}