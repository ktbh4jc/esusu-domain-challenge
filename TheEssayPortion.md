# Prompts: 
3. Explain in as much detail as you like how you would scale this API service to ultimately support a volume of 10,000 requests per second. Some things to consider include:  
a. How will you handle CI/CD in the context of a live service?  
b. How will you model and support SLAs? What should operational SLAs be for this service?  
c. How do you support geographically diverse clients? As you scale the system out horizontally, how do you continue to keep track of tokens without slowing down the system?

4. The success of this product has led to the team adding new and exciting features which
the customers have highly requested. Specifically, we are now offering a premium
offering: Memes AI. With Memes AI, you can get even spicier memes curated by
generative AI. Naturally, this feature costs extra money and requires a separate
subscription.
Describe how you would modify the service to now keep track of whether a client is authorized
to get AI-generated memes. If a client has this subscription, then they should get AI-memes,
and they should get normal memes otherwise. How do you keep track of authorization of a
client as we scale the system without slowing down performance?

# Response

My primary experience with Go has been in the context of writing lambdas connected by SQS queues. My initial thought process would be to see if we could split things out even further to make that work. The benefit provided here is that you're almost never wasting time waiting. Lambdas can be spun up as needed and are effectively infinitely scalable. SQS is also a horizontally scalable queueing system. At my previous company, I was working on a fairly intensive communication service hosted in lambdas that would send about a million emails a day with no signs of struggling. 

I also think that the lambda approach, when paired with dependency inversion, could pretty cleanly solve problem 4. One of the lambdas could fetch the whole user object, including a new `big_spender` (or realistically something like `ai_enabled`) flag. Then we could have two sqs queues hooked up to receive from that lambda, one that goes to the default meme maker, and another that goes to the much slower ai meme maker. 

Now that I've given the elevator pitch, time to answer the questions more specifically.  
- How will you handle CI/CD in the context of a live service?  
  Step 1: Ask an expert. I don't want them to solve the problem for me, but CI/CD is one of the areas where I am more of a consumer than a producer so if we have an existing pattern that we are happy with, I'm not going out of my way to rock that particular boat. General questions are things like "What are we doing currently?", "Do we have docs?", and "What do you think could have done better?"  
  Step 2: Ask the team. The expert and I could come up with something fantastic, but if nobody likes using it we might as well not reinvent the wheel.  
  Step 3: Act. Odds are, I am going to stick with what we already have unless there are any glaring issues. 

  You'll notice I haven't really name dropped any specific technologies here. My previous company used Jenkins for most of our CI/CD stuff and we owned our own deployments so I have become comfortable using it. Even wrote a few jenkins jobs through the magic of copy/paste/tweak. But there are a lot of tools out there and I don't have a strong enough expertise or opinion to stick to what I have used before. 

- How will you model and support SLAs? What should operational SLAs be for this service?  
  Main thoughts on SLAs relate to data retention, avoiding double charging, and content filtering. 
  - Data retention: Maybe something along the lines of "We will keep a log of any purchases and token spending, which users can access on our website." This particular service doesn't really store any user specific data, but if we did we would also need to have a policy around data removal. 
  - Double Charging: In order to avoid race conditions or double meme making, I would suggest that as part of the initial database reading we would also lock that user until the process is done. That way if someone tries to get two memes at once we will see that lock on the second and only spin up one meme. We would also be sure that we don't get in trouble when an "At least once" queue hits the lambda that removes a token twice. Suggested SLA here would be something along the lines of "We will use only one token for each meme generated. Meme generation can take up to [x] seconds." When in doubt, give the user a free meme so we don't get sued. 
  - Content Filtering: This is arguably more of a marketing thing, but  I suspect we would want to have some sort of content filter in place to not randomly generate anything the user didn't agree to. We could do this by having a set pool of templates that are pre-screened to be below a general level of risque. This gets trickier for AI as I would assume we are using an image generator. In this instance I say do what we can, but defer when we have to. Suggested SLA would be something like "MaaS has the following content policy [...]. If something generated on the platform violates this policy, reach out to [contact info] and one of our reps will get back to you in [x] days."

- How do you support geographically diverse clients? As you scale the system out horizontally, how do you continue to keep track of tokens without slowing down the system?  
  One of the benefits of the AWS approach is that all elements are horizontally scalable and AWS has hosting across the globe. Although connecting multiple AWS hosts is something I would need to look into the logistics of. One trick that comes to mind is to have a mid level queue of pending db changes as I _feel_ like that's where our troubles are going to live. But this also feels like a problem that has been solved enough times that I feel like we could devote a bit of time to research and get a first draft pretty quickly. 