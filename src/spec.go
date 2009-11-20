package specify

import(
	"fmt";
	"container/list";
	"os";
)

var behaviors *list.List;
var exitCode int;

func init() {
	behaviors = list.New();
}

type matcher interface {
	Be(interface{});
}

type should struct {
	value interface{};
}

func (self should) Be(value interface{}) {
	if self.value != value {
		self.Failf("Expected %v to be %v\n", self.value, value);
	}
}

func (self should) Failf(format string, v ...) {
	fmt.Printf(format, v);
	exitCode = 1;
}

type That struct {
	Should matcher;
}

type Expect interface {
	That(interface {}) *That;
}

type It interface {
	Should(string, func(*Expect));
}

type expect struct {}

func (self expect) That(value interface {}) (result *That) {
	result = &That{};
	result.Should = &should{value};
	return;
}

type it struct {
	name string;
	expectations *list.List;
}

func (self it) Should(desc string, spec func(*Expect)) {
	var e Expect = expect{};
	self.expectations.PushBack(func() { spec(&e); });
}

func (self it) run() {
	iter := self.expectations.Iter();
	for !closed(iter) {
		item := <-iter;
		if item == nil { break; }
		test,_ := item.(func());
		test();
	}
}

func Behavior(item string, spec func(*It)) {
	var i It = it{item, list.New()};
	spec(&i);
	behaviors.PushBack(i);
}

func Run() {
	iter := behaviors.Iter();
	for !closed(iter) {
		item := <-iter;
		if item == nil { break; }
		i,_ := item.(it);
		i.run();
	}
	os.Exit(exitCode);
}